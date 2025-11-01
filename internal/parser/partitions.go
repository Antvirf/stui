package parser

import (
	"encoding/json"
	"fmt"
	"time"

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

	// State & flags
	if apiPart.State != nil && len(*apiPart.State) > 0 {
		part.State = (*apiPart.State)[0]
	}
	if apiPart.Flags != nil {
		part.Flags = *apiPart.Flags
	}

	// Node allocation
	if apiPart.Nodes != nil && apiPart.Nodes.Configured != nil {
		part.Nodes = *apiPart.Nodes.Configured
	}
	if apiPart.NodeSets != nil {
		part.NodeCount = len(*apiPart.NodeSets)
	}
	if apiPart.Nodes != nil && apiPart.Nodes.Total != nil {
		part.TotalNodes = int(*apiPart.Nodes.Total)
	}
	if apiPart.Cpus != nil && apiPart.Cpus.Total != nil {
		part.TotalCPUs = int(*apiPart.Cpus.Total)
	}

	// Time limits
	if apiPart.Timeouts != nil {
		if apiPart.Timeouts.DefaultTime != nil && apiPart.Timeouts.DefaultTime.Number != nil {
			part.DefaultTime = time.Duration(*apiPart.Timeouts.DefaultTime.Number) * time.Minute
		}
		if apiPart.Timeouts.MaximumTime != nil && apiPart.Timeouts.MaximumTime.Number != nil {
			part.MaxTime = time.Duration(*apiPart.Timeouts.MaximumTime.Number) * time.Minute
		}
	}

	// Priority & scheduling
	if apiPart.Priority != nil && apiPart.Priority.JobFactor != nil {
		part.Priority = int(*apiPart.Priority.JobFactor)
	}
	if apiPart.Priority != nil && apiPart.Priority.Tier != nil {
		part.PriorityTier = int(*apiPart.Priority.Tier)
	}
	if apiPart.Oversubscribe != nil && len(*apiPart.Oversubscribe) > 0 {
		part.OverSubscribe = (*apiPart.Oversubscribe)[0]
	}
	if apiPart.Preemption != nil {
		if apiPart.Preemption.Type != nil && len(*apiPart.Preemption.Type) > 0 {
			part.Preempt = (*apiPart.Preemption.Type)[0]
		}
		if apiPart.Preemption.GraceTime != nil {
			part.GraceTime = int(*apiPart.Preemption.GraceTime)
		}
	}

	// Resource limits
	if apiPart.Maximums != nil {
		if apiPart.Maximums.CpusPerNode != nil && apiPart.Maximums.CpusPerNode.Number != nil {
			part.MaxCPUsPerNode = int(*apiPart.Maximums.CpusPerNode.Number)
		}
		if apiPart.Maximums.MemoryPerNode != nil {
			part.MaxMemPerNode = *apiPart.Maximums.MemoryPerNode
		}
		if apiPart.Maximums.MemoryPerCpu != nil {
			part.MaxMemPerCPU = *apiPart.Maximums.MemoryPerCpu
		}
		if apiPart.Maximums.Nodes != nil && apiPart.Maximums.Nodes.Number != nil {
			part.MaxNodes = int(*apiPart.Maximums.Nodes.Number)
		}
	}
	if apiPart.Minimums != nil {
		if apiPart.Minimums.Nodes != nil {
			part.MinNodes = int(*apiPart.Minimums.Nodes)
		}
	}
	if apiPart.Defaults != nil {
		if apiPart.Defaults.CpusPerNode != nil && apiPart.Defaults.CpusPerNode.Number != nil {
			part.DefaultCPUsPerNode = int(*apiPart.Defaults.CpusPerNode.Number)
		}
		if apiPart.Defaults.MemoryPerNode != nil {
			part.DefaultMemPerNode = *apiPart.Defaults.MemoryPerNode
		}
	}

	// Access control
	if apiPart.Accounts != nil {
		if apiPart.Accounts.Allow != nil {
			part.AllowAccounts = *apiPart.Accounts.Allow
		}
		if apiPart.Accounts.Deny != nil {
			part.DenyAccounts = *apiPart.Accounts.Deny
		}
	}
	if apiPart.Groups != nil && apiPart.Groups.Allowed != nil {
		part.AllowGroups = *apiPart.Groups.Allowed
	}
	if apiPart.Qos != nil {
		if apiPart.Qos.Allowed != nil {
			part.AllowQOS = *apiPart.Qos.Allowed
		}
		if apiPart.Qos.Denied != nil {
			part.DenyQOS = *apiPart.Qos.Denied
		}
		if apiPart.Qos.Assigned != nil {
			part.QOS = *apiPart.Qos.Assigned
		}
	}

	// Misc
	if apiPart.TresBillingWeights != nil {
		part.Billing = *apiPart.TresBillingWeights
	}
	if apiPart.AllowedAllocatingNodes != nil {
		part.Features = *apiPart.AllowedAllocatingNodes
	}

	return part
}
