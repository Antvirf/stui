package model

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

func GetNodesWithTimeout(timeout time.Duration, debugMultiplier int) (TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sinfo", "--Node", "--noheader", "-o=%N|%P|%T|%c|%m|%L|%E|%f|%F|%G|%X|%Y|%Z")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return TableData{}, fmt.Errorf("sinfo failed: %v", err)
	}

	headers := []string{
		"Node", "Partition", "State", "CPUs", "Memory",
		"CPULoad", "Reason", "Sockets", "Cores", "Threads", "GRES",
	}
	var rows [][]string

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) >= 11 {
			// Multiply each row according to DebugMultiplier
			for i := 0; i < debugMultiplier; i++ {
				nodeName := strings.TrimPrefix(fields[0], "=")
				if debugMultiplier > 1 {
					nodeName = fmt.Sprintf("%s-%d", nodeName, i+1)
				}
				row := []string{
					nodeName,   // Node
					fields[1],  // Partition
					fields[2],  // State
					fields[3],  // CPUs
					fields[4],  // Memory
					fields[5],  // CPULoad
					fields[6],  // Reason
					fields[7],  // Sockets
					fields[8],  // Cores
					fields[9],  // Threads
					fields[10], // GRES
				}
				rows = append(rows, row)
			}
		}
	}

	return TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func GetJobsWithTimeout(timeout time.Duration, debugMultiplier int) (TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "squeue", "--noheader", "-o=%i|%u|%P|%j|%T|%M|%N")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return TableData{}, fmt.Errorf("squeue failed: %v", err)
	}

	headers := []string{"ID", "User", "Partition", "Name", "State", "Time", "Nodes"}
	var rows [][]string

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) >= 7 {
			// Multiply each row according to DebugMultiplier
			for i := 0; i < debugMultiplier; i++ {
				jobID := strings.TrimPrefix(fields[0], "=")
				if debugMultiplier > 1 {
					jobID = fmt.Sprintf("%s-%d", jobID, i+1)
				}
				row := []string{
					jobID,                              // Job ID
					strings.TrimPrefix(fields[1], "="), // User
					strings.TrimPrefix(fields[2], "="), // Partition
					strings.TrimPrefix(fields[3], "="), // Name
					strings.TrimPrefix(fields[4], "="), // State
					strings.TrimPrefix(fields[5], "="), // Time
					strings.TrimPrefix(fields[6], "="), // Nodes
				}
				rows = append(rows, row)
			}
		}
	}

	return TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}

func GetNodeDetailsWithTimeout(nodeName string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scontrol", "show", "node", nodeName)
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

	cmd := exec.CommandContext(ctx, "scontrol", "show", "job", baseJobID)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		return "", fmt.Errorf("scontrol failed: %v", err)
	}
	return string(out), nil
}

func GetSchedulerInfo() (string, string) {
	// Get scheduler host from slurm config
	cmd := exec.Command("scontrol", "show", "config")
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

	cmd := exec.CommandContext(ctx, "sdiag")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		return "", fmt.Errorf("sdiag failed: %v", err)
	}

	return string(out), nil
}
