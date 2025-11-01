package domain

import "time"

// Job represents a Slurm job in domain terms, independent of display or API format
type Job struct {
	// Core identification
	JobID       string
	ArrayJobID  string
	ArrayTaskID string
	Name        string

	// Ownership & accounting
	User      string
	Account   string
	Partition string
	QOS       string

	// State & lifecycle
	State          string
	StateReason    string
	ExitCode       string
	DerivedExitCode string

	// Time tracking
	SubmitTime    *time.Time
	StartTime     *time.Time
	EndTime       *time.Time
	EligibleTime  *time.Time
	RunTime       time.Duration
	TimeLimit     time.Duration
	Deadline      *time.Time

	// Resources
	Nodes         string
	NodeCount     int
	CPUs          int
	Memory        string
	TRES          string
	UsedGRES      string

	// Execution details
	WorkingDir    string
	Command       string
	Script        string
	StdOut        string
	StdErr        string
	StdIn         string
	Comment       string

	// Priority & constraints
	Priority    int
	Constraints string
	Features    string

	// Misc
	Licenses     string
	Cluster      string
	Reservation  string
	FailedNode   string
	Container    string
	Flags        []string
}

// Jobs is a collection of Job entities
type Jobs []Job
