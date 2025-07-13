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

func getSacctMgrDataWithTimeout(command string, timeout time.Duration, columns *[]config.ColumnConfig, computeColumnWidths bool) (*TableData, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "sacctmgr") + " " + command

	cmd := exec.CommandContext(ctx,
		strings.Split(fullCommand, " ")[0],
		strings.Split(fullCommand, " ")[1:]...,
	)

	if config.SacctMgrCurrentEntity == "RunAwayJobs" {
		// For RunAwayJobs, we need to input an "N" as the command is interactive
		// and the interactivity cannot be disabled.
		stdIn, _ := cmd.StdinPipe()
		stdIn.Write([]byte("no"))
		defer stdIn.Close()
	}

	rawOut, err := cmd.CombinedOutput()
	out := string(rawOut)
	execTime := time.Since(startTime).Milliseconds()

	// Runawayjobs always prints something to stderr, so we need to check if the output is an actual error
	if config.SacctMgrCurrentEntity != "RunAwayJobs" {
		if strings.HasPrefix(out, "NOTE: ") { // This signifies it's OK, in that case we nil the error.
			err = nil
		}
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("sacctmgr: timed out after %dms: %s", execTime, fullCommand)
			return EmptyTableData(), fmt.Errorf("timeout after %v", timeout)
		}

		logger.Debugf("sacctmgr: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return EmptyTableData(), fmt.Errorf("%v", out)
	}

	logger.Debugf("sacctmgr: completed in %dms: %s", execTime, fullCommand)

	rawRows := []map[string]string{}
	if config.SacctMgrCurrentEntity == "RunAwayJobs" {
		rawRows = parseSacctMgrRunawayJobsOutput(out)
	} else {
		rawRows = parseSacctOutput(out)
	}

	var rows [][]string
	for _, rawRow := range rawRows {
		// Each row will have all of its fields, no filtering
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

			row[j] = safeGetFromMap(rawRow, col.DisplayName)
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers:             columns,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}, nil
}
