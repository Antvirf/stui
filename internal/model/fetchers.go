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

func GetNodesWithTimeout(timeout time.Duration, debugMultiplier int) (*TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "node",
	)

	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("scontrol failed: %v", err)
	}

	// Parse the output into node entries
	nodes := parseScontrolOutput("NodeName=", string(out))

	// Get configured columns
	columns := strings.Split(config.NodeViewColumns, ",")
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col // Use raw field names as headers initially
	}

	var rows [][]string
	for _, node := range nodes {
		for i := 0; i < debugMultiplier; i++ {
			nodeName := node["NodeName"]
			if debugMultiplier > 1 {
				nodeName = fmt.Sprintf("%s-%d", nodeName, i+1)
			}

			row := make([]string, len(columns))
			for j, col := range columns {
				if col == "NodeName" {
					row[j] = nodeName
				} else {
					row[j] = safeGetFromMap(node, col)
				}
			}
			rows = append(rows, row)
		}
	}

	return &TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func GetJobsWithTimeout(timeout time.Duration, debugMultiplier int) (*TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "job",
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("scontrol failed: %v", err)
	}

	// Parse the output into job entries
	jobs := parseScontrolOutput("JobId=", string(out))

	// Get configured columns
	columns := strings.Split(config.JobViewColumns, ",")
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col // Use raw field names as headers initially
	}

	var rows [][]string
	for _, job := range jobs {
		for i := 0; i < debugMultiplier; i++ {
			jobID := job["JobId"]
			if debugMultiplier > 1 {
				jobID = fmt.Sprintf("%s-%d", jobID, i+1)
			}

			row := make([]string, len(columns))
			for j, col := range columns {
				if col == "JobId" {
					row[j] = jobID
				} else {
					row[j] = safeGetFromMap(job, col)
				}
			}
			rows = append(rows, row)
		}
	}

	return &TableData{
		Headers: headers,
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
	// Strip any debug multiplier suffix
	baseJobID := jobID
	if idx := strings.LastIndex(jobID, "-"); idx != -1 {
		baseJobID = jobID[:idx]
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "scontrol"),
		"show", "job", baseJobID,
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
