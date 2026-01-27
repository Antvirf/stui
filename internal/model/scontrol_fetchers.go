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

func getScontrolDataWithTimeout(command string, columns *[]config.ColumnConfig, timeout time.Duration, computeColumnWidths bool, parserFunction func(string) []map[string]string) (*TableData, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "scontrol") + " " + command
	cmd := execStringCommand(ctx, fullCommand)
	rawOut, err := cmd.CombinedOutput()
	out := string(rawOut)
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return EmptyTableData(), fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return EmptyTableData(), fmt.Errorf("%v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)

	rawRows := parserFunction(out)

	var rows [][]CellValue
	for _, rawRow := range rawRows {
		row := make([]CellValue, len(*columns))
		for j := range *columns {
			// Access elements by index so we modify the original
			col := &(*columns)[j]

			// Handle computed columns (ratios)
			if col.ComputedColumn && col.ComputationType == "ratio" {
				// Parse components as typed values for ratio computation
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
				// Normal single column
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
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}, nil
}

func GetNodeDetailsWithTimeout(nodeName string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := fmt.Sprintf("%s show node %s", path.Join(config.SlurmBinariesPath, "scontrol"), nodeName)
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func GetJobDetailsWithTimeout(jobID string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := fmt.Sprintf("%s show job %s", path.Join(config.SlurmBinariesPath, "scontrol"), jobID)
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func getSdiagWithTimeout(timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "sdiag")
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("sdiag: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %dms", timeout)
		}
		logger.Debugf("sdiag: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("sdiag failed: %v", err)
	}

	logger.Debugf("sdiag: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func execStringCommand(ctx context.Context, cmd string) *exec.Cmd {
	return exec.CommandContext(ctx, strings.Split(cmd, " ")[0], strings.Split(cmd, " ")[1:]...)
}
