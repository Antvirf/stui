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

// Check whether the cluster is reacheable with 'scontrol ping'
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
