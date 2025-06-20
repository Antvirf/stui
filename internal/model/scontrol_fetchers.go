package model

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

func getScontrolDataWithTimeout(command string, columns *[]config.ColumnConfig, timeout time.Duration) (*TableData, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "scontrol") + " " + command
	cmd := execStringCommand(ctx, fullCommand)
	rawOut, err := cmd.CombinedOutput()
	out := string(rawOut)
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return EmptyTableData(), fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return EmptyTableData(), fmt.Errorf("%v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)

	rawRows := parseScontrolOutput(out)

	var rows [][]string
	for _, rawRow := range rawRows {
		row := make([]string, len(*columns))
		for j, col := range *columns {
			if col.DividedByColumn {
				components := strings.Split(col.Name, "//")
				var values []string
				for _, component := range components {
					values = append(values, safeGetFromMap(rawRow, component))
				}
				row[j] = strings.Join(values, " / ")
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
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := fmt.Sprintf("%s show node %s", path.Join(config.SlurmBinariesPath, "scontrol"), nodeName)
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func GetJobDetailsWithTimeout(jobID string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := fmt.Sprintf("%s show job %s", path.Join(config.SlurmBinariesPath, "scontrol"), jobID)
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("scontrol: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("scontrol: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("scontrol failed: %v", err)
	}

	logger.Debugf("scontrol: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func GetSacctJobDetailsWithTimeout(jobID string, timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := fmt.Sprintf(
		"%s -j %s --format %s --parsable",
		path.Join(config.SlurmBinariesPath, "sacct"),
		jobID,
		config.AllSacctViewColumns,
	)
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("sacct: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %v", timeout)
		}
		logger.Debugf("sacct: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("sacct failed: %v", err)
	}

	logger.Debugf("sacct: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func getSdiagWithTimeout(timeout time.Duration) (string, error) {
	startTime := time.Now()
	FetchCounter.increment()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fullCommand := path.Join(config.SlurmBinariesPath, "sdiag")
	cmd := execStringCommand(ctx, fullCommand)
	out, err := cmd.Output()
	execTime := time.Since(startTime).Milliseconds()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debugf("sdiag: timed out after %dms: %s", execTime, fullCommand)
			return "", fmt.Errorf("timeout after %dms", timeout)
		}
		logger.Debugf("sdiag: failed after %dms: %s (%v)", execTime, fullCommand, err)
		return "", fmt.Errorf("sdiag failed: %v", err)
	}

	logger.Debugf("sdiag: completed in %dms: %s", execTime, fullCommand)
	return string(out), nil
}

func execStringCommand(ctx context.Context, cmd string) *exec.Cmd {
	return exec.CommandContext(ctx, strings.Split(cmd, " ")[0], strings.Split(cmd, " ")[1:]...)
}
