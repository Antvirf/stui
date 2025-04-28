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

func getScontrolDataWithTimeout(command string, columns *[]config.ColumnConfig, timeout time.Duration) (*TableData, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		strings.Split(command, " ")...,
	)
	rawOut, err := cmd.CombinedOutput()
	out := string(rawOut)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("%v", err)
	}

	rawRows := parseScontrolOutput(out)

	var rows [][]string
	for _, rawRow := range rawRows {
		row := make([]string, len(*columns))
		for j, col := range *columns {
			if col.DividedByColumn {
				components := strings.Split(col.Name, "//")
				var values []string
				for _, component := range components {
					values = append(values, safeGetFromMap(rawRow, component))
				}
				row[j] = strings.Join(values, " / ")
			} else {
				// Normal cell
				row[j] = safeGetFromMap(rawRow, col.Name)
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: columns,
		Rows:    rows,
	}, nil
}

func GetNodeDetailsWithTimeout(nodeName string, timeout time.Duration) (string, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "node", nodeName,
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		return "", fmt.Errorf("scontrol failed: %v", err)
	}
	return string(out), nil
}

func GetJobDetailsWithTimeout(jobID string, timeout time.Duration) (string, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "job", jobID,
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		return "", fmt.Errorf("scontrol failed: %v", err)
	}
	return string(out), nil
}

func getSdiagWithTimeout(timeout time.Duration) (string, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "sdiag"),
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		return "", fmt.Errorf("sdiag failed: %v", err)
	}

	return string(out), nil
}
