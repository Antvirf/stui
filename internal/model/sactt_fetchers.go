package model

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

func GetSacctMgrEntityWithTimeout(entity string, timeout time.Duration) (*TableData, error) {
	// TODO: In the future we may want to make this column config also configurable.
	// However, since there are so many different sacct entities, we probably
	// need to support configuration files to make this usable.
	// For now, we hardcode the column maps to useful defaults.

	var columns []config.ColumnConfig
	columnConfig := strings.Split(SACCTMGR_ENTITY_COLUMN_CONFIGS[entity], ",")
	for _, key := range columnConfig {
		columns = append(columns, config.ColumnConfig{Name: key})
	}

	data, err := getSacctMgrDataWithTimeout(
		fmt.Sprintf("show %s --parsable2", entity),
		config.RequestTimeout,
		&columns,
	)
	return data, err
}

func getSacctMgrDataWithTimeout(command string, timeout time.Duration, columns *[]config.ColumnConfig) (*TableData, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "sacctmgr"),
		strings.Split(command, " ")...,
	)
	rawOut, err := cmd.CombinedOutput()
	out := string(rawOut)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("%v", out)
	}

	rawRows := parseSacctOutput(out)

	var rows [][]string
	for _, rawRow := range rawRows {
		// Each row will have all of its fields, no filtering
		row := make([]string, len(*columns))
		for j, col := range *columns {
			row[j] = safeGetFromMap(rawRow, col.Name)
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: columns,
		Rows:    rows,
	}, nil
}
