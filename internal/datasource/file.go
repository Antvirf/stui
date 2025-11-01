package datasource

import (
	"context"
	"fmt"
	"os"
)

// FileSource implements SlurmDataSource by reading JSON from files
// Useful for testing and offline development
type FileSource struct {
	jobsFile       string
	nodesFile      string
	partitionsFile string
}

// NewFileSource creates a new file-based data source
func NewFileSource(jobsFile, nodesFile, partitionsFile string) *FileSource {
	return &FileSource{
		jobsFile:       jobsFile,
		nodesFile:      nodesFile,
		partitionsFile: partitionsFile,
	}
}

// Name returns the name of this data source
func (f *FileSource) Name() string {
	return "FileJSON"
}

// FetchJobsJSON reads jobs JSON from file
func (f *FileSource) FetchJobsJSON(ctx context.Context) ([]byte, error) {
	if f.jobsFile == "" {
		return nil, fmt.Errorf("jobs file path not configured")
	}
	data, err := os.ReadFile(f.jobsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read jobs file: %w", err)
	}
	return data, nil
}

// FetchNodesJSON reads nodes JSON from file
func (f *FileSource) FetchNodesJSON(ctx context.Context) ([]byte, error) {
	if f.nodesFile == "" {
		return nil, fmt.Errorf("nodes file path not configured")
	}
	data, err := os.ReadFile(f.nodesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read nodes file: %w", err)
	}
	return data, nil
}

// FetchPartitionsJSON reads partitions JSON from file
func (f *FileSource) FetchPartitionsJSON(ctx context.Context) ([]byte, error) {
	if f.partitionsFile == "" {
		return nil, fmt.Errorf("partitions file path not configured")
	}
	data, err := os.ReadFile(f.partitionsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions file: %w", err)
	}
	return data, nil
}

// FetchJobDetailJSON reads job detail - for files, just returns same as FetchJobsJSON
func (f *FileSource) FetchJobDetailJSON(ctx context.Context, jobID string) ([]byte, error) {
	return f.FetchJobsJSON(ctx)
}

// FetchNodeDetailJSON reads node detail - for files, just returns same as FetchNodesJSON
func (f *FileSource) FetchNodeDetailJSON(ctx context.Context, nodeName string) ([]byte, error) {
	return f.FetchNodesJSON(ctx)
}
