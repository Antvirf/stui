package config

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"
)

// Check whether we should enable sacctmgr-related features, by
// making a call to 'sacctmgr show configuration' and checking its
// exit code.
func checkIfSacctMgrIsAvailable() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(SlurmBinariesPath, "sacctmgr"),
		strings.Split("show cluster", " ")...,
	)
	_, err := cmd.Output()
	if err != nil {
		SacctEnabled = false
	} else {
		SacctEnabled = true
	}
}

// Check whether the cluster is reachable with 'scontrol ping'
func checkIfClusterIsReachable() error {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		path.Join(SlurmBinariesPath, "scontrol"), "ping",
	)
	rawOut, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("cluster did not respond within configured timeout %s", RequestTimeout)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to Slurm: %s", string(rawOut))
	}
	return nil
}

// Fetch scheduler info with a timeout. This used to be in fetchers,
// but is a one-off call at initialization and makes more sense
// in checks.
func getSchedulerInfoWithTimeout(timeout time.Duration) (schedulerHostName, clusterName, slurmVersion string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx,
		path.Join(SlurmBinariesPath, "scontrol"),
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

		if strings.HasPrefix(line, "SLURM_VERSION") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				// Extract host from ClusterName = name
				slurmVersion = strings.TrimSpace(parts[1])
			}
		}
	}
	return host, clusterName, slurmVersion
}
