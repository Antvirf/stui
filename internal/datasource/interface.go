package datasource

import (
	"context"
)

// SlurmDataSource is the interface for fetching JSON data from various sources
// This abstraction allows for:
// - Binary calls to Slurm commands (scontrol, sacct, etc.) with --json
// - REST API calls to slurmrestd
// - File-based sources for testing/offline mode
type SlurmDataSource interface {
	// FetchJobsJSON retrieves jobs data as JSON
	FetchJobsJSON(ctx context.Context) ([]byte, error)

	// FetchNodesJSON retrieves nodes data as JSON
	FetchNodesJSON(ctx context.Context) ([]byte, error)

	// FetchPartitionsJSON retrieves partitions data as JSON
	FetchPartitionsJSON(ctx context.Context) ([]byte, error)

	// FetchJobDetailJSON retrieves detailed info for a specific job as JSON
	FetchJobDetailJSON(ctx context.Context, jobID string) ([]byte, error)

	// FetchNodeDetailJSON retrieves detailed info for a specific node as JSON
	FetchNodeDetailJSON(ctx context.Context, nodeName string) ([]byte, error)

	// Name returns the name/type of this data source for logging
	Name() string
}
