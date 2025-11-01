package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/domain"
	"github.com/antvirf/stui/internal/slurmapi"
)

// ParseNodesJSON parses JSON from scontrol show node --json into domain.Nodes
func ParseNodesJSON(jsonData []byte) (domain.Nodes, error) {
	var resp slurmapi.V0043OpenapiNodesResp
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes JSON: %w", err)
	}

	nodes := make(domain.Nodes, 0, len(resp.Nodes))
	for _, apiNode := range resp.Nodes {
		node := convertNodeToDomain(&apiNode)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// convertNodeToDomain converts slurmapi.V0043Node to domain.Node
func convertNodeToDomain(apiNode *slurmapi.V0043Node) domain.Node {
	node := domain.Node{}

	// Core identification
	if apiNode.Name != nil {
		node.Name = *apiNode.Name
	}
	if apiNode.Address != nil {
		node.Address = *apiNode.Address
	}
	if apiNode.Hostname != nil {
		node.Hostname = *apiNode.Hostname
	}
	if apiNode.ClusterName != nil {
		node.Cluster = *apiNode.ClusterName
	}

	// State & status
	if apiNode.State != nil {
		node.State = *apiNode.State
	}
	if apiNode.Reason != nil {
		node.StateReason = *apiNode.Reason
	}
	if apiNode.ReasonSetByUser != nil {
		node.ReasonSetByUser = *apiNode.ReasonSetByUser
	}
	if apiNode.ReasonChangedAt != nil && apiNode.ReasonChangedAt.Number != nil {
		t := time.Unix(int64(*apiNode.ReasonChangedAt.Number), 0)
		node.ReasonChangedAt = &t
	}
	if apiNode.NextStateAfterReboot != nil {
		node.NextStateAfterReboot = *apiNode.NextStateAfterReboot
	}

	// Hardware specs
	if apiNode.Architecture != nil {
		node.Architecture = *apiNode.Architecture
	}
	if apiNode.Cpus != nil {
		node.CPUs = int(*apiNode.Cpus)
	}
	if apiNode.EffectiveCpus != nil {
		node.EffectiveCPUs = int(*apiNode.EffectiveCpus)
	}
	if apiNode.SpecializedCpus != nil {
		node.SpecializedCPUs = *apiNode.SpecializedCpus
	}
	if apiNode.Boards != nil {
		node.Boards = int(*apiNode.Boards)
	}
	if apiNode.Cores != nil {
		node.Cores = int(*apiNode.Cores)
	}
	if apiNode.SpecializedCores != nil {
		node.SpecializedCores = int(*apiNode.SpecializedCores)
	}
	if apiNode.Sockets != nil {
		node.Sockets = int(*apiNode.Sockets)
	}
	if apiNode.Threads != nil {
		node.Threads = int(*apiNode.Threads)
	}

	// Memory
	if apiNode.RealMemory != nil {
		node.RealMemory = *apiNode.RealMemory
	}
	if apiNode.AllocMemory != nil {
		node.AllocMemory = *apiNode.AllocMemory
	}
	if apiNode.FreeMem != nil && apiNode.FreeMem.Number != nil {
		node.FreeMemory = int64(*apiNode.FreeMem.Number)
	}
	if apiNode.SpecializedMemory != nil {
		node.SpecializedMemory = *apiNode.SpecializedMemory
	}

	// Resources
	if apiNode.Gres != nil {
		node.GRES = *apiNode.Gres
	}
	if apiNode.GresDrained != nil {
		node.GRESDrained = *apiNode.GresDrained
	}
	if apiNode.GresUsed != nil {
		node.GRESUsed = *apiNode.GresUsed
	}
	if apiNode.GpuSpec != nil {
		node.GPUSpec = *apiNode.GpuSpec
	}

	// Features & constraints
	if apiNode.Features != nil {
		node.Features = *apiNode.Features
	}
	if apiNode.ActiveFeatures != nil {
		node.ActiveFeatures = *apiNode.ActiveFeatures
	}
	if apiNode.Extra != nil {
		node.Extra = *apiNode.Extra
	}

	// Partitions & ownership
	if apiNode.Partitions != nil {
		node.Partitions = *apiNode.Partitions
	}
	if apiNode.Owner != nil {
		node.Owner = *apiNode.Owner
	}

	// System info
	if apiNode.OperatingSystem != nil {
		node.OperatingSystem = *apiNode.OperatingSystem
	}
	if apiNode.BootTime != nil && apiNode.BootTime.Number != nil {
		t := time.Unix(int64(*apiNode.BootTime.Number), 0)
		node.BootTime = &t
	}
	if apiNode.LastBusy != nil && apiNode.LastBusy.Number != nil {
		t := time.Unix(int64(*apiNode.LastBusy.Number), 0)
		node.LastBusy = &t
	}
	if apiNode.CpuLoad != nil {
		node.CPULoad = int(*apiNode.CpuLoad)
	}

	// Network
	if apiNode.Port != nil {
		node.Port = int(*apiNode.Port)
	}
	if apiNode.BurstbufferNetworkAddress != nil {
		node.BurstBufferNetworkAddress = *apiNode.BurstbufferNetworkAddress
	}

	// Cloud
	if apiNode.InstanceId != nil {
		node.InstanceID = *apiNode.InstanceId
	}
	if apiNode.InstanceType != nil {
		node.InstanceType = *apiNode.InstanceType
	}

	// Misc
	if apiNode.Comment != nil {
		node.Comment = *apiNode.Comment
	}
	if apiNode.McsLabel != nil {
		node.MCSLabel = *apiNode.McsLabel
	}
	if apiNode.Weight != nil {
		node.Weight = int(*apiNode.Weight)
	}

	return node
}

// formatNodeState converts node state array to string
func formatNodeState(states *[]string) string {
	if states == nil || len(*states) == 0 {
		return ""
	}
	return strings.Join(*states, "+")
}
