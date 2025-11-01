package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/domain"
)

// JobsToTableData converts domain.Jobs to TableData for display
func JobsToTableData(jobs domain.Jobs, columns *[]config.ColumnConfig) *TableData {
	rows := make([][]string, 0, len(jobs))
	
	for _, job := range jobs {
		row := make([]string, len(*columns))
		
		for j, col := range *columns {
			row[j] = extractJobField(job, col.DisplayName)
		}
		
		rows = append(rows, row)
	}

	return &TableData{
		Headers:             columns,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}
}

// extractJobField extracts a field from domain.Job based on field name
func extractJobField(job domain.Job, fieldName string) string {
	switch fieldName {
	case "JobID":
		return job.JobID
	case "ArrayJobId":
		return job.ArrayJobID
	case "ArrayTaskId":
		return job.ArrayTaskID
	case "Name", "JobName":
		return job.Name
	case "UserId", "User", "UserName":
		return job.User
	case "Account":
		return job.Account
	case "Partition":
		return job.Partition
	case "QOS":
		return job.QOS
	case "JobState", "State":
		return job.State
	case "StateReason", "Reason":
		return job.StateReason
	case "ExitCode":
		return job.ExitCode
	case "DerivedExitCode":
		return job.DerivedExitCode
	case "SubmitTime":
		if job.SubmitTime != nil {
			return job.SubmitTime.Format("2006-01-02T15:04:05")
		}
		return ""
	case "StartTime":
		if job.StartTime != nil {
			return job.StartTime.Format("2006-01-02T15:04:05")
		}
		return ""
	case "EndTime":
		if job.EndTime != nil {
			return job.EndTime.Format("2006-01-02T15:04:05")
		}
		return ""
	case "EligibleTime":
		if job.EligibleTime != nil {
			return job.EligibleTime.Format("2006-01-02T15:04:05")
		}
		return ""
	case "RunTime":
		return formatDuration(job.RunTime)
	case "TimeLimit":
		return formatDuration(job.TimeLimit)
	case "Nodes", "NodeList":
		return job.Nodes
	case "NodeCount", "NumNodes":
		return strconv.Itoa(job.NodeCount)
	case "Cpus", "NumCPUs":
		return strconv.Itoa(job.CPUs)
	case "Memory":
		return job.Memory
	case "TresAllocStr", "TRES":
		return job.TRES
	case "UsedGres", "Gres":
		return job.UsedGRES
	case "WorkDir", "WorkingDirectory":
		return job.WorkingDir
	case "Command":
		return job.Command
	case "StdOut", "StandardOutput":
		return job.StdOut
	case "StdErr", "StandardError":
		return job.StdErr
	case "StdIn", "StandardInput":
		return job.StdIn
	case "Comment":
		return job.Comment
	case "Priority":
		return strconv.Itoa(job.Priority)
	case "Features", "Constraints":
		return job.Constraints
	case "Licenses":
		return job.Licenses
	case "Cluster":
		return job.Cluster
	case "Reservation":
		return job.Reservation
	case "FailedNode":
		return job.FailedNode
	case "Container":
		return job.Container
	default:
		return ""
	}
}

// NodesToTableData converts domain.Nodes to TableData for display
func NodesToTableData(nodes domain.Nodes, columns *[]config.ColumnConfig) *TableData {
	rows := make([][]string, 0, len(nodes))
	
	for _, node := range nodes {
		row := make([]string, len(*columns))
		
		for j, col := range *columns {
			row[j] = extractNodeField(node, col.DisplayName)
		}
		
		rows = append(rows, row)
	}

	return &TableData{
		Headers:             columns,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}
}

// extractNodeField extracts a field from domain.Node based on field name
func extractNodeField(node domain.Node, fieldName string) string {
	switch fieldName {
	case "NodeName", "Name":
		return node.Name
	case "NodeAddr", "Address":
		return node.Address
	case "NodeHostName", "Hostname":
		return node.Hostname
	case "Cluster":
		return node.Cluster
	case "State":
		return strings.Join(node.State, "+")
	case "StateReason", "Reason":
		return node.StateReason
	case "ReasonSetByUser":
		return node.ReasonSetByUser
	case "Architecture", "Arch":
		return node.Architecture
	case "CPUs", "Cpus":
		return strconv.Itoa(node.CPUs)
	case "EffectiveCPUs":
		return strconv.Itoa(node.EffectiveCPUs)
	case "SpecializedCPUs":
		return node.SpecializedCPUs
	case "Boards":
		return strconv.Itoa(node.Boards)
	case "Cores":
		return strconv.Itoa(node.Cores)
	case "SpecializedCores":
		return strconv.Itoa(node.SpecializedCores)
	case "Sockets":
		return strconv.Itoa(node.Sockets)
	case "Threads":
		return strconv.Itoa(node.Threads)
	case "RealMemory", "Memory":
		return strconv.FormatInt(node.RealMemory, 10)
	case "AllocMem":
		return strconv.FormatInt(node.AllocMemory, 10)
	case "FreeMem":
		return strconv.FormatInt(node.FreeMemory, 10)
	case "SpecializedMemory":
		return strconv.FormatInt(node.SpecializedMemory, 10)
	case "Gres", "GRES":
		return node.GRES
	case "GRESDrained":
		return node.GRESDrained
	case "GRESUsed":
		return node.GRESUsed
	case "GPUSpec":
		return node.GPUSpec
	case "Features", "AvailableFeatures":
		return strings.Join(node.Features, ",")
	case "ActiveFeatures":
		return strings.Join(node.ActiveFeatures, ",")
	case "Partitions":
		return strings.Join(node.Partitions, ",")
	case "Owner":
		return node.Owner
	case "OS", "OperatingSystem":
		return node.OperatingSystem
	case "BootTime":
		if node.BootTime != nil {
			return node.BootTime.Format("2006-01-02T15:04:05")
		}
		return ""
	case "CPULoad":
		return strconv.Itoa(node.CPULoad)
	case "Port":
		return strconv.Itoa(node.Port)
	case "Comment":
		return node.Comment
	case "Weight":
		return strconv.Itoa(node.Weight)
	default:
		return ""
	}
}

// PartitionsToTableData converts domain.Partitions to TableData for display
func PartitionsToTableData(partitions domain.Partitions, columns *[]config.ColumnConfig) *TableData {
	rows := make([][]string, 0, len(partitions))
	
	for _, partition := range partitions {
		row := make([]string, len(*columns))
		
		for j, col := range *columns {
			row[j] = extractPartitionField(partition, col.DisplayName)
		}
		
		rows = append(rows, row)
	}

	return &TableData{
		Headers:             columns,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}
}

// extractPartitionField extracts a field from domain.Partition based on field name
func extractPartitionField(partition domain.Partition, fieldName string) string {
	switch fieldName {
	case "PartitionName", "Name":
		return partition.Name
	case "Cluster":
		return partition.Cluster
	case "State":
		return partition.State
	case "Nodes":
		return partition.Nodes
	case "TotalNodes":
		return strconv.Itoa(partition.TotalNodes)
	case "TotalCPUs":
		return strconv.Itoa(partition.TotalCPUs)
	case "DefaultTime":
		return formatDuration(partition.DefaultTime)
	case "MaxTime":
		return formatDuration(partition.MaxTime)
	case "Priority":
		return strconv.Itoa(partition.Priority)
	case "PriorityTier":
		return strconv.Itoa(partition.PriorityTier)
	case "OverSubscribe":
		return partition.OverSubscribe
	case "Preempt":
		return partition.Preempt
	case "GraceTime":
		return strconv.Itoa(partition.GraceTime)
	case "MaxCPUsPerNode":
		return strconv.Itoa(partition.MaxCPUsPerNode)
	case "MaxMemPerNode":
		return strconv.FormatInt(partition.MaxMemPerNode, 10)
	case "MinNodes":
		return strconv.Itoa(partition.MinNodes)
	case "MaxNodes":
		return strconv.Itoa(partition.MaxNodes)
	case "AllowAccounts":
		return strings.Join(partition.AllowAccounts, ",")
	case "DenyAccounts":
		return strings.Join(partition.DenyAccounts, ",")
	case "AllowGroups":
		return strings.Join(partition.AllowGroups, ",")
	case "AllowQOS":
		return strings.Join(partition.AllowQOS, ",")
	case "QOS":
		return partition.QOS
	case "TRES":
		return partition.TRES
	default:
		return ""
	}
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
