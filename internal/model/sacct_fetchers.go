package model

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

func GetSacctData(since time.Duration) (*TableData, error) {
	cmd := exec.Command(
		path.Join(config.SlurmBinariesPath, "sacct"),
		"--allusers",
		"--long",
		"--allocations",
		"--parsable2",
		fmt.Sprintf("--starttime=now-%d", int(since.Seconds())),
	)

	rawOut, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("sacct failed: %v\nOutput: %s", err, string(rawOut))
	}

	return parseSacctOutputToTableData(string(rawOut))
}

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
