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

	fullCommand := fmt.Sprintf("%s --allusers --allocations --parsable2 --delimiter %s --starttime=now-%d --format %s",
		path.Join(config.SlurmBinariesPath, "sacct"),
		SACCT_DELIMITER,
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
			return EmptyTableDataWithColumns(columns), fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("sacct: failed out after %dms: %s", execTime, fullCommand)
		return EmptyTableDataWithColumns(columns), fmt.Errorf("%v", err)
	}

	logger.Debugf("sacct: completed in %dms: %s", execTime, fullCommand)
	return parseSacctOutputToTableData(out, columns, computeColumnWidths)
}
func parseSacctOutputToTableData(output string, columns *[]config.ColumnConfig, computeColumnWidths bool) (*TableData, error) {
	rawRows := parseSacctOutput(output, SACCT_DELIMITER)
	if len(rawRows) == 0 {
		return EmptyTableDataWithColumns(columns), nil
	}

	var rows [][]CellValue
	for _, rawRow := range rawRows {
		row := make([]CellValue, len(*columns))
		for j := range *columns {
			// Access elements by index so we modify the original
			col := &(*columns)[j]

			// Handle computed columns (ratios)
			if col.ComputedColumn && col.ComputationType == "ratio" {
				// RawName is normalized to use // separator for compatibility with GetColumnFields
				components := strings.Split(col.RawName, "//")
				if len(components) != 2 {
					// Invalid format, treat as string
					row[j] = NewStringValue("ERROR: Invalid ratio format")
					continue
				}

				// Get field types for both components
				numeratorField := strings.TrimSpace(components[0])
				denominatorField := strings.TrimSpace(components[1])

				numeratorType := GetFieldType(numeratorField)
				denominatorType := GetFieldType(denominatorField)

				// Parse raw values as typed
				numeratorRaw := safeGetFromMap(rawRow, numeratorField)
				denominatorRaw := safeGetFromMap(rawRow, denominatorField)

				numeratorVal := parseTypedValue(numeratorRaw, numeratorType)
				denominatorVal := parseTypedValue(denominatorRaw, denominatorType)

				// Create ratio value
				row[j] = NewRatioValue(numeratorVal, denominatorVal)

			} else if col.DividedByColumn {
				// Non-computed divided columns (just display side-by-side)
				components := strings.Split(col.RawName, "//")
				var values []string
				for _, component := range components {
					values = append(values, safeGetFromMap(rawRow, component))
				}
				rawValue := strings.Join(values, " / ")
				row[j] = parseTypedValue(rawValue, TypeString)

			} else {
				// Normal single column - use type lookup
				rawValue := safeGetFromMap(rawRow, col.DisplayName)
				fieldType := GetFieldType(col.DisplayName)
				row[j] = parseTypedValue(rawValue, fieldType)
			}

			// Update column width based on display value
			if computeColumnWidths {
				displayLen := len(row[j].Display())
				col.Width = min(
					max(displayLen, col.Width),
					config.MaximumColumnWidth,
				)
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers:             columns,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows, columns),
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
		"%s -j %s --format %s --parsable --delimiter %s",
		path.Join(config.SlurmBinariesPath, "sacct"),
		jobID,
		columnStrings,
		SACCT_DELIMITER,
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
