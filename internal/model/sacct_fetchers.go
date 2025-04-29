package model

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
)

// GetSacctDataFunc is the function signature for getting sacct data
type GetSacctDataFunc func(since time.Duration) (*TableData, error)

// RealGetSacctData is the actual implementation of GetSacctData
func RealGetSacctData(since time.Duration) (*TableData, error) {
	startTime := time.Now()
	FetchCounter.increment()

	fullCommand := fmt.Sprintf("%s --allusers --long --allocations --parsable2 --starttime=now-%d",
		path.Join(config.SlurmBinariesPath, "sacct"),
		max(
			int(config.RefreshInterval.Seconds()),
			int(since.Seconds()),
			1,
		))

	cmd := exec.Command(
		strings.Split(fullCommand, " ")[0],
		strings.Split(fullCommand, " ")[1:]...,
	)

	rawOut, err := cmd.CombinedOutput()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		logger.Debugf("sacct: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return nil, fmt.Errorf("sacct failed: %v\nOutput: %s", err, string(rawOut))
	}

	logger.Debugf("sacct: completed in %dms: %s", execTime, fullCommand)
	return parseSacctOutputToTableData(string(rawOut))
}

// GetSacctData is a variable that points to the function to use for getting sacct data.
// This allows us to mock it in tests.
var GetSacctData GetSacctDataFunc = RealGetSacctData

func parseSacctOutputToTableData(output string) (*TableData, error) {
	entries := parseSacctOutput(output)
	if len(entries) == 0 {
		return &TableData{
			Headers: &[]config.ColumnConfig{},
			Rows:    [][]string{},
		}, nil
	}

	var headers []config.ColumnConfig
	columnConfig := strings.Split(SACCT_COLUMNS, ",")
	for _, key := range columnConfig {
		headers = append(headers, config.ColumnConfig{Name: key})
	}

	var rows [][]string
	for _, entry := range entries {
		row := make([]string, len(headers))
		for i, col := range headers {
			row[i] = safeGetFromMap(entry, col.Name)
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: &headers,
		Rows:    rows,
	}, nil
}
