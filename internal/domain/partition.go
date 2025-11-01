package domain

import "time"

// Partition represents a Slurm partition in domain terms
type Partition struct {
	// Core identification
	Name    string
	Cluster string

	// State & flags
	State string
	Flags []string

	// Node allocation
	Nodes          string
	NodeCount      int
	TotalNodes     int
	TotalCPUs      int
	AllocatedNodes int
	AllocatedCPUs  int

	// Time limits
	DefaultTime time.Duration
	MaxTime     time.Duration

	// Priority & scheduling
	Priority           int
	PriorityTier       int
	OverSubscribe      string
	Preempt            string
	GraceTime          int
	PreemptMode        string

	// Resource limits
	MaxCPUsPerNode    int
	MaxMemPerNode     int64
	MaxMemPerCPU      int64
	MinNodes          int
	MaxNodes          int
	DefaultCPUsPerNode int
	DefaultMemPerNode int64

	// Access control
	AllowAccounts  []string
	DenyAccounts   []string
	AllowGroups    []string
	AllowQOS       []string
	DenyQOS        []string
	QOS            string

	// Misc
	TRES        string
	Billing     string
	Features    string
	MaxJobsAccrue int
}

// Partitions is a collection of Partition entities
type Partitions []Partition
