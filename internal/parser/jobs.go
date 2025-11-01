package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/domain"
	"github.com/antvirf/stui/internal/slurmapi"
)

// ParseJobsJSON parses JSON from scontrol show job --json into domain.Jobs
func ParseJobsJSON(jsonData []byte) (domain.Jobs, error) {
	var resp slurmapi.V0043OpenapiJobInfoResp
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jobs JSON: %w", err)
	}

	jobs := make(domain.Jobs, 0, len(resp.Jobs))
	for _, apiJob := range resp.Jobs {
		job := convertJobInfoToDomain(&apiJob)
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// convertJobInfoToDomain converts slurmapi.V0043JobInfo to domain.Job
func convertJobInfoToDomain(apiJob *slurmapi.V0043JobInfo) domain.Job {
	job := domain.Job{}

	// Core identification
	if apiJob.JobId != nil {
		job.JobID = strconv.Itoa(int(*apiJob.JobId))
	}
	if apiJob.ArrayJobId != nil && apiJob.ArrayJobId.Number != nil {
		job.ArrayJobID = strconv.Itoa(int(*apiJob.ArrayJobId.Number))
	}
	if apiJob.ArrayTaskId != nil && apiJob.ArrayTaskId.Number != nil {
		job.ArrayTaskID = strconv.Itoa(int(*apiJob.ArrayTaskId.Number))
	}
	if apiJob.Name != nil {
		job.Name = *apiJob.Name
	}

	// Ownership & accounting
	if apiJob.Account != nil {
		job.Account = *apiJob.Account
	}
	if apiJob.UserId != nil {
		job.User = strconv.Itoa(int(*apiJob.UserId))
	}
	if apiJob.UserName != nil {
		job.User = *apiJob.UserName
	}
	if apiJob.Partition != nil {
		job.Partition = *apiJob.Partition
	}
	if apiJob.Qos != nil {
		job.QOS = *apiJob.Qos
	}

	// State & lifecycle
	if apiJob.JobState != nil && len(*apiJob.JobState) > 0 {
		job.State = (*apiJob.JobState)[0]
	}
	if apiJob.StateReason != nil {
		job.StateReason = *apiJob.StateReason
	}
	if apiJob.ExitCode != nil {
		job.ExitCode = formatExitCode(apiJob.ExitCode)
	}
	if apiJob.DerivedExitCode != nil {
		job.DerivedExitCode = formatExitCode(apiJob.DerivedExitCode)
	}

	// Time tracking
	if apiJob.SubmitTime != nil && apiJob.SubmitTime.Number != nil {
		t := time.Unix(int64(*apiJob.SubmitTime.Number), 0)
		job.SubmitTime = &t
	}
	if apiJob.StartTime != nil && apiJob.StartTime.Number != nil {
		t := time.Unix(int64(*apiJob.StartTime.Number), 0)
		job.StartTime = &t
	}
	if apiJob.EndTime != nil && apiJob.EndTime.Number != nil {
		t := time.Unix(int64(*apiJob.EndTime.Number), 0)
		job.EndTime = &t
	}
	if apiJob.EligibleTime != nil && apiJob.EligibleTime.Number != nil {
		t := time.Unix(int64(*apiJob.EligibleTime.Number), 0)
		job.EligibleTime = &t
	}
	if apiJob.RunTime != nil && apiJob.RunTime.Number != nil {
		job.RunTime = time.Duration(*apiJob.RunTime.Number) * time.Second
	}
	if apiJob.TimeLimit != nil && apiJob.TimeLimit.Number != nil {
		job.TimeLimit = time.Duration(*apiJob.TimeLimit.Number) * time.Minute
	}
	if apiJob.Deadline != nil && apiJob.Deadline.Number != nil {
		t := time.Unix(int64(*apiJob.Deadline.Number), 0)
		job.Deadline = &t
	}

	// Resources
	if apiJob.Nodes != nil {
		job.Nodes = *apiJob.Nodes
	}
	if apiJob.NodeCount != nil {
		job.NodeCount = int(*apiJob.NodeCount)
	}
	if apiJob.Cpus != nil && apiJob.Cpus.Number != nil {
		job.CPUs = int(*apiJob.Cpus.Number)
	}
	if apiJob.Memory != nil {
		job.Memory = *apiJob.Memory
	}
	if apiJob.TresAllocStr != nil {
		job.TRES = *apiJob.TresAllocStr
	}
	if apiJob.UsedGres != nil {
		job.UsedGRES = *apiJob.UsedGres
	}

	// Execution details
	if apiJob.WorkDir != nil {
		job.WorkingDir = *apiJob.WorkDir
	}
	if apiJob.Command != nil {
		job.Command = *apiJob.Command
	}
	if apiJob.BatchScript != nil {
		job.Script = *apiJob.BatchScript
	}
	if apiJob.StandardOutput != nil {
		job.StdOut = *apiJob.StandardOutput
	}
	if apiJob.StandardError != nil {
		job.StdErr = *apiJob.StandardError
	}
	if apiJob.StandardInput != nil {
		job.StdIn = *apiJob.StandardInput
	}
	if apiJob.Comment != nil {
		job.Comment = *apiJob.Comment
	}

	// Priority & constraints
	if apiJob.Priority != nil && apiJob.Priority.Number != nil {
		job.Priority = int(*apiJob.Priority.Number)
	}
	if apiJob.Features != nil {
		job.Constraints = *apiJob.Features
		job.Features = *apiJob.Features
	}

	// Misc
	if apiJob.Licenses != nil {
		job.Licenses = *apiJob.Licenses
	}
	if apiJob.Cluster != nil {
		job.Cluster = *apiJob.Cluster
	}
	if apiJob.Reservation != nil {
		job.Reservation = *apiJob.Reservation
	}
	if apiJob.FailedNode != nil {
		job.FailedNode = *apiJob.FailedNode
	}
	if apiJob.Container != nil {
		job.Container = *apiJob.Container
	}
	if apiJob.Flags != nil {
		job.Flags = *apiJob.Flags
	}

	return job
}

// formatExitCode converts exit code struct to string representation
func formatExitCode(ec *slurmapi.V0043ProcessExitCodeVerbose) string {
	if ec == nil {
		return ""
	}
	if ec.Status != nil && len(*ec.Status) > 0 {
		return (*ec.Status)[0]
	}
	if ec.ReturnCode != nil && ec.ReturnCode.Number != nil {
		return strconv.Itoa(int(*ec.ReturnCode.Number))
	}
	return ""
}

// safeStringJoin safely joins string slices, handling nil pointers
func safeStringJoin(slice *[]string, sep string) string {
	if slice == nil || len(*slice) == 0 {
		return ""
	}
	return strings.Join(*slice, sep)
}
