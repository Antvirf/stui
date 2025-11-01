package datasource

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
)

// BinaryJSONSource implements SlurmDataSource by calling Slurm binaries with --json flag
type BinaryJSONSource struct {
	timeout time.Duration
}

// NewBinaryJSONSource creates a new binary-based data source
func NewBinaryJSONSource(timeout time.Duration) *BinaryJSONSource {
	return &BinaryJSONSource{
		timeout: timeout,
	}
}

// Name returns the name of this data source
func (b *BinaryJSONSource) Name() string {
	return "BinaryJSON"
}

// FetchJobsJSON fetches jobs using scontrol with JSON output
func (b *BinaryJSONSource) FetchJobsJSON(ctx context.Context) ([]byte, error) {
	startTime := time.Now()
	
	cmdCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "scontrol") + " show job --detail --all --json"
	cmd := b.execStringCommand(cmdCtx, fullCommand)
	
	rawOut, err := cmd.CombinedOutput()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return nil, fmt.Errorf("timeout after %v", b.timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return nil, fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return rawOut, nil
}

// FetchNodesJSON fetches nodes using scontrol with JSON output
func (b *BinaryJSONSource) FetchNodesJSON(ctx context.Context) ([]byte, error) {
	startTime := time.Now()
	
	cmdCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "scontrol") + " show node --detail --all --json"
	cmd := b.execStringCommand(cmdCtx, fullCommand)
	
	rawOut, err := cmd.CombinedOutput()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return nil, fmt.Errorf("timeout after %v", b.timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return nil, fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return rawOut, nil
}

// FetchPartitionsJSON fetches partitions using scontrol with JSON output
func (b *BinaryJSONSource) FetchPartitionsJSON(ctx context.Context) ([]byte, error) {
	startTime := time.Now()
	
	cmdCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "scontrol") + " show partition --all --json"
	cmd := b.execStringCommand(cmdCtx, fullCommand)
	
	rawOut, err := cmd.CombinedOutput()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return nil, fmt.Errorf("timeout after %v", b.timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return nil, fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return rawOut, nil
}

// FetchJobDetailJSON fetches detailed job info using scontrol with JSON output
func (b *BinaryJSONSource) FetchJobDetailJSON(ctx context.Context, jobID string) ([]byte, error) {
	startTime := time.Now()
	
	cmdCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	fullCommand := fmt.Sprintf("%s show job %s --json", path.Join(config.SlurmBinariesPath, "scontrol"), jobID)
	cmd := b.execStringCommand(cmdCtx, fullCommand)
	
	rawOut, err := cmd.CombinedOutput()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return nil, fmt.Errorf("timeout after %v", b.timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return nil, fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return rawOut, nil
}

// FetchNodeDetailJSON fetches detailed node info using scontrol with JSON output
func (b *BinaryJSONSource) FetchNodeDetailJSON(ctx context.Context, nodeName string) ([]byte, error) {
	startTime := time.Now()
	
	cmdCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	fullCommand := fmt.Sprintf("%s show node %s --json", path.Join(config.SlurmBinariesPath, "scontrol"), nodeName)
	cmd := b.execStringCommand(cmdCtx, fullCommand)
	
	rawOut, err := cmd.CombinedOutput()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return nil, fmt.Errorf("timeout after %v", b.timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return nil, fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return rawOut, nil
}

// execStringCommand is a helper to execute a command string
func (b *BinaryJSONSource) execStringCommand(ctx context.Context, cmd string) *exec.Cmd {
	parts := strings.Split(cmd, " ")
	return exec.CommandContext(ctx, parts[0], parts[1:]...)
}
