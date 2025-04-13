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

	SACCTMGR_ENTITY_COLUMN_CONFIGS = map[string]string{
		"Account":     "Org,Account,Descr",
		"Association": "Cluster,Account,User,Partition,Share,QOS,Def QOS,Priority,GrpJobs,GrpTRES,GrpSubmit,GrpWall,GrpTRESMins,MaxJobs,MaxTRES,MaxTRESPerNode,MaxSubmit,MaxWall,MaxTRESMins,GrpTRESRunMins",
		"Cluster":     "Cluster,ControlHost,ControlPort,RPC,Share,QOS,Def QOS,GrpJobs,GrpTRES,GrpSubmit,MaxJobs,MaxTRES,MaxSubmit,MaxWall",
		"Event":       "Cluster,NodeName,TimeStart,TimeEnd,State,Reason,User",
		"Federation":  "Cluster,NodeName,TimeStart,TimeEnd,State,Reason,User",
		"QOS":         "Name,Priority,GraceTime,Preempt,PreemptExemptTime,PreemptMode,Flags,UsageThres,UsageFactor,GrpTRES,GrpTRESMins,GrpTRESRunMins,GrpJobs,GrpSubmit,GrpWall,MaxTRES,MaxTRESPerNode,MaxTRESMins,MaxWall,MaxTRESPU,MaxJobsPU,MaxSubmitPU,MaxTRESPA,MaxTRESRunMinsPA,MaxTRESRunMinsPU,MaxJobsPA,MaxSubmitPA,MinTRES",
		"Resource":    "Name,Server,Type,Count,LastConsumed,Allocated,ServerType,Flags",
		"Reservation": "Cluster,Name,TRES,TimeStart,TimeEnd,UnusedWall",
		"Transaction": "Time,Action,Actor,Where,Info",
		"TRES":        "Type,Name,ID",
		"User":        "User, Def Acct, Def WCKey, Admin",
	}
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
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("sacctmgr failed: %v", err)
	}

	rawRows := parseSacctOutput(string(out))

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
