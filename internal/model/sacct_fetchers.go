package model

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
)

func getSacctDataSinceWithTimeout(since time.Duration, columns *[]config.ColumnConfig, timeout time.Duration) (*TableData, error) {
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
		logger.Debugf("sacct: failed after %dms: %s (%v)", execTime, fullCommand, err)
		log.Fatalf("sacct: timed out after %dms (its timeout setting is %d times the standard request timeout): %s", execTime, config.SacctTimeoutMultiplier, fullCommand)
	}

	logger.Debugf("sacct: completed in %dms: %s", execTime, fullCommand)
	return parseSacctOutputToTableData(out, columns)
}
func parseSacctOutputToTableData(output string, columns *[]config.ColumnConfig) (*TableData, error) {
	entries := parseSacctOutput(output)
	if len(entries) == 0 {
		return &TableData{
			Headers: &[]config.ColumnConfig{},
			Rows:    [][]string{},
		}, nil
	}

	var rows [][]string
	for _, entry := range entries {
		row := make([]string, len(*columns))
		for i, col := range *columns {
			// Check if it's a combined column
			if strings.Contains(col.Name, "//") {
				parts := strings.Split(col.Name, "//")
				combinedValue := ""
				for j, part := range parts {
					if j > 0 {
						combinedValue += " / "
					}
					combinedValue += safeGetFromMap(entry, part)
				}
				row[i] = combinedValue
			} else {
				row[i] = safeGetFromMap(entry, col.Name)
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: columns,
		Rows:    rows,
	}, nil
}
