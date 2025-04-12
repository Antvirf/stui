package model

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

var (
	SACCTMGR_TABLE_ENTITIES = []string{
		"Account",
		"Association",
		"Cluster",
		"Event",
		"Federation",
		// "Problem", // Requires admin
		"QOS",
		"Resource",
		"Reservation",
		// "RunAwayJobs", // Requires operator perms
		"Transaction",
		"TRES",
		"User",
		// "WCKey", // Requires admin
	}
	SACCTMGR_TEXT_ENTITIES = []string{
		"Configuration",
		// "Stats", // Requires admin
	}
	// SACCMGR_ENTITIES = append(SACCTMGR_TABLE_ENTITIES, SACCTMGR_TEXT_ENTITIES...)
)

func GetSacctMgrEntityWithTimeout(entity string, timeout time.Duration) (*TableData, error) {
	data, err := getSacctMgrDataWithTimeout(
		fmt.Sprintf("show %s --parsable2", entity),
		config.RequestTimeout,
	)
	return data, err
}

func getSacctMgrDataWithTimeout(command string, timeout time.Duration) (*TableData, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "sacctmgr"),
		strings.Split(command, " ")...,
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("sacctmgr failed: %v", err)
	}

	rawRows := parseSacctOutput(string(out))

	// TODO: In the future we may want to make this column config also configurable.
	// However, since there are so many different sacct entities, we probably
	// need to support configuration files to make this usable.
	// For now, we create the columns map on-demand, and just render everything.
	var columns []config.ColumnConfig
	if len(rawRows) > 0 {
		for key := range rawRows[0] {
			columns = append(columns, config.ColumnConfig{Name: key})
		}
	}

	// Sort columns by alphabetic order of Name
	// We need to do this, because we do not enforce a specific column list.
	// Maps are unordered, so columns will jump around.
	sort.Slice(columns, func(i, j int) bool {
		return columns[i].Name < columns[j].Name
	})

	var rows [][]string
	for _, rawRow := range rawRows {
		// Each row will have all of its fields, no filtering
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = safeGetFromMap(rawRow, col.Name)
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: &columns,
		Rows:    rows,
	}, nil
}
