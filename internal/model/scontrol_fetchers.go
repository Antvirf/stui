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

func getScontrolDataWithTimeout(command string, columns *[]config.ColumnConfig, partitionFilter string, prefix string, timeout time.Duration) (*TableData, error) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		strings.Split(command, " ")...,
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("scontrol failed: %v", err)
	}

	rawRows := parseScontrolOutput(prefix, string(out))

	// Depending on data, the partition field name may be called differently. Deal with both cases.
	partitionFieldname := "Partition"
	if len(rawRows) != 0 {
		if _, exists := rawRows[0]["Partitions"]; exists {
			partitionFieldname = "Partitions"
		}
	}

	var rows [][]string
	for _, rawRow := range rawRows {
		// Apply partition filter if set
		if partitionFilter != "" {
			if !strings.Contains(rawRow[partitionFieldname], partitionFilter) {
				continue
			}
		}

		row := make([]string, len(*columns))
		for j, col := range *columns {
			if col.DividedByColumn {
				components := strings.Split(col.Name, "//")
				top := safeGetFromMap(rawRow, components[0])
				bottom := safeGetFromMap(rawRow, components[1])
				row[j] = fmt.Sprintf("%s / %s", top, bottom)

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

func GetSchedulerInfoWithTimeout(timeout time.Duration) (schedulerHostName, clusterName string) {
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "config",
	)
	out, err := cmd.Output()
	if err != nil {
		schedulerHostName = "(failed to fetch scheduler info)"
	}

	// Parse output for controller host
	var host string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "SlurmctldHost") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				// Extract host from SlurmctldHost[0]=hostname
				host = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "ClusterName") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				// Extract host from ClusterName = name
				clusterName = strings.TrimSpace(parts[1])
			}
		}
	}
	return host, clusterName
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
