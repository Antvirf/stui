package config

import (
	"context"
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
