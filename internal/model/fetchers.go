package model

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

func GetNodesWithTimeout(timeout time.Duration) (*TableData, error) {
	// Prep columns
	columns := strings.Split(config.NodeViewColumns, ",")
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col // Use raw field names as headers initially
	}
	data, err := GetScontrolDataWithTimeout(
		"show node --detail --all",
		columns,
		config.PartitionFilter,
		"NodeName=",
		config.RequestTimeout,
	)
	return data, err
}

func GetJobsWithTimeout(timeout time.Duration) (*TableData, error) {
	// Prep columns
	columns := strings.Split(config.JobViewColumns, ",")
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col // Use raw field names as headers initially
	}
	data, err := GetScontrolDataWithTimeout(
		"show job --detail --all",
		columns,
		config.PartitionFilter,
		"JobId=",
		config.RequestTimeout,
	)
	return data, err
}

func GetScontrolDataWithTimeout(command string, columns []string, partitionFilter string, prefix string, timeout time.Duration) (*TableData, error) {
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
	if _, exists := rawRows[0]["Partitions"]; exists {
		partitionFieldname = "Partitions"
	}

	var rows [][]string
	for _, rawRow := range rawRows {
		// Apply partition filter if set
		if config.PartitionFilter != "" {
			matched := false
			filteredPartitions := strings.Split(config.PartitionFilter, ",")
			for _, partition := range filteredPartitions {
				if strings.Contains(rawRow[partitionFieldname], partition) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = safeGetFromMap(rawRow, col)
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: columns,
		Rows:    rows,
	}, nil
}

func GetNodeDetailsWithTimeout(nodeName string, timeout time.Duration) (string, error) {
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

func GetSchedulerInfoWithTimeout(timeout time.Duration) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "config",
	)
	out, err := cmd.Output()
	if err != nil {
		return "unknown", "unknown"
	}

	// Parse output for controller host
	var host string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "SlurmctldHost") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				// Extract host from SlurmctldHost[0]=hostname
				host = strings.TrimSpace(parts[1])
				if strings.Contains(host, "[") {
					host = strings.Split(host, "[")[0]
				}
				break
			}
		}
	}

	if host == "" {
		return "unknown", "unknown"
	}

	// Try to get IP
	addrs, err := net.LookupHost(host)
	if err == nil && len(addrs) > 0 {
		return host, addrs[0]
	}

	// Try short hostname if FQDN failed
	if strings.Contains(host, ".") {
		shortHost := strings.Split(host, ".")[0]
		addrs, err = net.LookupHost(shortHost)
		if err == nil && len(addrs) > 0 {
			return host, addrs[0]
		}
	}

	return host, "unknown"
}

func GetSdiagWithTimeout(timeout time.Duration) (string, error) {
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
