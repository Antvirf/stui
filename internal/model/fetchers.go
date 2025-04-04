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
	nodes := parseScontrolOutput(string(out))

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

			reason := "N/A"
			if r, ok := node["Reason"]; ok && r != "(null)" {
				reason = r
			}

			gres := "(null)"
			if g, ok := node["Gres"]; ok {
				gres = g
			}

			row := []string{
				nodeName,               // Node
				node["Partitions"],     // Partitions
				node["State"],          // State
				node["CPUTot"],         // CPUs
				node["RealMemory"],     // Memory
				node["CPULoad"],        // CPULoad
				reason,                 // Reason
				node["Sockets"],        // Sockets
				node["CoresPerSocket"], // Cores
				node["ThreadsPerCore"], // Threads
				gres,                   // GRES
			}
			rows = append(rows, row)
		}
	}

	return &TableData{
		Headers: headers,
		Rows:    rows,
	}, nil
}

// parseScontrolOutput parses the scontrol show node output into a slice of maps
func parseScontrolOutput(output string) []map[string]string {
	var nodes []map[string]string
	currentNode := make(map[string]string)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new node entry
		if strings.HasPrefix(line, "NodeName=") {
			if len(currentNode) > 0 {
				nodes = append(nodes, currentNode)
			}
			currentNode = make(map[string]string)
		}

		// Parse key=value pairs
		pairs := strings.Fields(line)
		for _, pair := range pairs {
			if idx := strings.Index(pair, "="); idx > 0 {
				key := pair[:idx]
				value := pair[idx+1:]
				currentNode[key] = value
			}
		}
	}

	// Add the last node if it exists
	if len(currentNode) > 0 {
		nodes = append(nodes, currentNode)
	}

	return nodes
}

func GetJobsWithTimeout(timeout time.Duration, debugMultiplier int) (*TableData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(config.SlurmBinariesPath, "squeue"),
		"--noheader", "-o=%i|%u|%P|%j|%T|%M|%N",
	)
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &TableData{}, fmt.Errorf("timeout after %v", timeout)
		}
		return &TableData{}, fmt.Errorf("squeue failed: %v", err)
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
