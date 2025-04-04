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

	headers := []string{
		"Node", "Partitions", "State", "CPUs", "Memory",
		"CPULoad", "Reason", "Sockets", "Cores", "Threads", "GRES",
	}
	var rows [][]string

	for _, node := range nodes {
		// Multiply each row according to DebugMultiplier
		for i := 0; i < debugMultiplier; i++ {
			nodeName := node["NodeName"]
			if debugMultiplier > 1 {
				nodeName = fmt.Sprintf("%s-%d", nodeName, i+1)
			}

			row := []string{
				nodeName, // Node
				safeGetFromMap(node, "Partitions"),
				safeGetFromMap(node, "State"),
				safeGetFromMap(node, "CPUTot"),
				safeGetFromMap(node, "RealMemory"),
				safeGetFromMap(node, "CPULoad"),
				safeGetFromMap(node, "Reason"),
				safeGetFromMap(node, "Sockets"),
				safeGetFromMap(node, "CoresPerSocket"),
				safeGetFromMap(node, "ThreadsPerCore"),
				safeGetFromMap(node, "Gres"),
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

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	var rows [][]string

	for _, job := range jobs {
		// Multiply each row according to DebugMultiplier
		for i := 0; i < debugMultiplier; i++ {
			jobID := job["JobId"]
			if debugMultiplier > 1 {
				jobID = fmt.Sprintf("%s-%d", jobID, i+1)
			}

			row := []string{
				jobID, // Job ID
				safeGetFromMap(job, "UserId"),
				safeGetFromMap(job, "Partition"),
				safeGetFromMap(job, "JobName"),
				safeGetFromMap(job, "JobState"),
				safeGetFromMap(job, "RunTime"),
				safeGetFromMap(job, "NodeList"),
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
