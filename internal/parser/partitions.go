package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antvirf/stui/internal/domain"
	"github.com/antvirf/stui/internal/slurmapi"
)

// ParsePartitionsJSON parses JSON from scontrol show partition --json into domain.Partitions
func ParsePartitionsJSON(jsonData []byte) (domain.Partitions, error) {
	var resp slurmapi.V0043OpenapiPartitionResp
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal partitions JSON: %w", err)
	}

	partitions := make(domain.Partitions, 0, len(resp.Partitions))
	for _, apiPartition := range resp.Partitions {
		partition := convertPartitionToDomain(&apiPartition)
		partitions = append(partitions, partition)
	}

	return partitions, nil
}

// convertPartitionToDomain converts slurmapi.V0043PartitionInfo to domain.Partition  
func convertPartitionToDomain(apiPart *slurmapi.V0043PartitionInfo) domain.Partition {
	part := domain.Partition{}

	// Core identification
	if apiPart.Name != nil {
		part.Name = *apiPart.Name
	}
	if apiPart.Cluster != nil {
		part.Cluster = *apiPart.Cluster
	}

	// State
	if apiPart.Partition != nil && apiPart.Partition.State != nil && len(*apiPart.Partition.State) > 0 {
		part.State = (*apiPart.Partition.State)[0]
	}

	// Nodes  
	if apiPart.Nodes != nil && apiPart.Nodes.Configured != nil {
		part.Nodes = *apiPart.Nodes.Configured
	}
	if apiPart.Nodes != nil && apiPart.Nodes.Total != nil {
		part.TotalNodes = int(*apiPart.Nodes.Total)
	}
	if apiPart.Cpus != nil && apiPart.Cpus.Total != nil {
		part.TotalCPUs = int(*apiPart.Cpus.Total)
	}

	// Priority
	if apiPart.Priority != nil && apiPart.Priority.JobFactor != nil {
		part.Priority = int(*apiPart.Priority.JobFactor)
	}
	if apiPart.Priority != nil && apiPart.Priority.Tier != nil {
		part.PriorityTier = int(*apiPart.Priority.Tier)
	}
	
	// Grace time
	if apiPart.GraceTime != nil {
		part.GraceTime = int(*apiPart.GraceTime)
	}

	// Limits
	if apiPart.Maximums != nil {
		if apiPart.Maximums.CpusPerNode != nil && apiPart.Maximums.CpusPerNode.Number != nil {
			part.MaxCPUsPerNode = int(*apiPart.Maximums.CpusPerNode.Number)
		}
		if apiPart.Maximums.MemoryPerCpu != nil {
			part.MaxMemPerCPU = *apiPart.Maximums.MemoryPerCpu
		}
		if apiPart.Maximums.Nodes != nil && apiPart.Maximums.Nodes.Number != nil {
			part.MaxNodes = int(*apiPart.Maximums.Nodes.Number)
		}
	}
	if apiPart.Minimums != nil && apiPart.Minimums.Nodes != nil {
		part.MinNodes = int(*apiPart.Minimums.Nodes)
	}

	// Accounts - these are comma-separated strings, not arrays
	if apiPart.Accounts != nil {
		if apiPart.Accounts.Allowed != nil && *apiPart.Accounts.Allowed != "" {
			part.AllowAccounts = strings.Split(*apiPart.Accounts.Allowed, ",")
		}
		if apiPart.Accounts.Deny != nil && *apiPart.Accounts.Deny != "" {
			part.DenyAccounts = strings.Split(*apiPart.Accounts.Deny, ",")
		}
	}
	
	// Groups - also comma-separated string
	if apiPart.Groups != nil && apiPart.Groups.Allowed != nil && *apiPart.Groups.Allowed != "" {
		part.AllowGroups = strings.Split(*apiPart.Groups.Allowed, ",")
	}

	// QOS - same pattern
	if apiPart.Qos != nil {
		if apiPart.Qos.Allowed != nil && *apiPart.Qos.Allowed != "" {
			part.AllowQOS = strings.Split(*apiPart.Qos.Allowed, ",")
		}
		if apiPart.Qos.Deny != nil && *apiPart.Qos.Deny != "" {
			part.DenyQOS = strings.Split(*apiPart.Qos.Deny, ",")
		}
		if apiPart.Qos.Assigned != nil {
			part.QOS = *apiPart.Qos.Assigned
		}
	}

	// TRES billing
	if apiPart.Tres != nil && apiPart.Tres.BillingWeights != nil {
		part.Billing = *apiPart.Tres.BillingWeights
	}

	return part
}
