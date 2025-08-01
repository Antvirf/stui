package model

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
)

func getSacctDataSinceWithTimeout(since time.Duration, columns *[]config.ColumnConfig, timeout time.Duration, computeColumnWidths bool) (*TableData, error) {
	startTime := time.Now()
	FetchCounter.increment()

	fullCommand := fmt.Sprintf("%s --allusers --allocations --parsable2 --starttime=now-%d --format %s",
		path.Join(config.SlurmBinariesPath, "sacct"),
		max(
			int(config.RefreshInterval.Seconds()),
			int(since.Seconds()),
			1,
		),
		strings.Join(config.GetColumnFields(columns), ","),
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		strings.Split(fullCommand, " ")[0],
		strings.Split(fullCommand, " ")[1:]...,
	)
	rawOut, err := cmd.CombinedOutput()
	out := string(rawOut)
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("sacct: timed out after %dms (its timeout setting is %d times the standard request timeout): %s", execTime, config.SacctTimeoutMultiplier, fullCommand)
			return EmptyTableData(), fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("sacct: failed out after %dms: %s", execTime, fullCommand)
		return EmptyTableData(), fmt.Errorf("%v", timeout)
	}

	logger.Debugf("sacct: completed in %dms: %s", execTime, fullCommand)
	return parseSacctOutputToTableData(out, columns, computeColumnWidths)
}
func parseSacctOutputToTableData(output string, columns *[]config.ColumnConfig, computeColumnWidths bool) (*TableData, error) {
	rawRows := parseSacctOutput(output)
	if len(rawRows) == 0 {
		return EmptyTableData(), nil
	}

	var rows [][]string
	for _, rawRow := range rawRows {
		row := make([]string, len(*columns))
		for j := range *columns {
			// Access elements by index so we modify the original
			col := &(*columns)[j]

			if computeColumnWidths {
				col.Width = min(
					max( // Increase col width if current cell is bigger than current max
						len(safeGetFromMap(rawRow, col.DisplayName)),
						col.Width,
					),
					config.MaximumColumnWidth, // .. but don't go above this value.
				)
			}

			// Check if it's a combined column
			if col.DividedByColumn {
				components := strings.Split(col.RawName, "//")
				var values []string
				for _, component := range components {
					values = append(values, safeGetFromMap(rawRow, component))
				}
				row[j] = strings.Join(values, " / ")
			} else {
				row[j] = safeGetFromMap(rawRow, col.DisplayName)
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: columns,
		Rows:    rows,
	}, nil
}

func GetSacctJobDetailsWithTimeout(jobID string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Clean up the columns list
	columnStrings := strings.ReplaceAll(config.AllSacctViewColumns, "++", "")
	columnStrings = strings.ReplaceAll(columnStrings, "//", ",")

	fullCommand := fmt.Sprintf(
		"%s -j %s --format %s --parsable",
		path.Join(config.SlurmBinariesPath, "sacct"),
		jobID,
		columnStrings,
	)
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("sacct: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("sacct: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("sacct failed: %v", err)
	}

	logger.Debugf("sacct: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}
