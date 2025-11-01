package domain

import "time"

// Node represents a Slurm compute node in domain terms
type Node struct {
	// Core identification
	Name     string
	Address  string
	Hostname string
	Cluster  string

	// State & status
	State            []string
	StateReason      string
	ReasonSetByUser  string
	ReasonChangedAt  *time.Time
	NextStateAfterReboot []string

	// Hardware specs
	Architecture      string
	CPUs              int
	EffectiveCPUs     int
	SpecializedCPUs   string
	Boards            int
	Cores             int
	SpecializedCores  int
	Sockets           int
	Threads           int

	// Memory
	RealMemory         int64
	AllocMemory        int64
	FreeMemory         int64
	SpecializedMemory  int64

	// Resources
	GRES        string
	GRESDrained string
	GRESUsed    string
	GPUSpec     string

	// Features & constraints
	Features       []string
	ActiveFeatures []string
	Extra          string

	// Partitions & ownership
	Partitions []string
	Owner      string

	// System info
	OperatingSystem string
	BootTime        *time.Time
	LastBusy        *time.Time
	CPULoad         int

	// Network
	Port                      int
	BurstBufferNetworkAddress string

	// Cloud
	InstanceID   string
	InstanceType string

	// Misc
	Comment  string
	MCSLabel string
	Weight   int
}

// Nodes is a collection of Node entities
type Nodes []Node
