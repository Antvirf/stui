package datasource

import (
	"context"
	"fmt"
)

// RestAPISource implements SlurmDataSource by calling slurmrestd REST API
// This is a placeholder for future implementation
type RestAPISource struct {
	baseURL string
	token   string
}

// NewRestAPISource creates a new REST API-based data source
func NewRestAPISource(baseURL, token string) *RestAPISource {
	return &RestAPISource{
		baseURL: baseURL,
		token:   token,
	}
}

// Name returns the name of this data source
func (r *RestAPISource) Name() string {
	return "RestAPI"
}

// FetchJobsJSON fetches jobs via slurmrestd REST API
func (r *RestAPISource) FetchJobsJSON(ctx context.Context) ([]byte, error) {
	return nil, fmt.Errorf("RestAPISource not yet implemented")
}

// FetchNodesJSON fetches nodes via slurmrestd REST API
func (r *RestAPISource) FetchNodesJSON(ctx context.Context) ([]byte, error) {
	return nil, fmt.Errorf("RestAPISource not yet implemented")
}

// FetchPartitionsJSON fetches partitions via slurmrestd REST API
func (r *RestAPISource) FetchPartitionsJSON(ctx context.Context) ([]byte, error) {
	return nil, fmt.Errorf("RestAPISource not yet implemented")
}

// FetchJobDetailJSON fetches job detail via slurmrestd REST API
func (r *RestAPISource) FetchJobDetailJSON(ctx context.Context, jobID string) ([]byte, error) {
	return nil, fmt.Errorf("RestAPISource not yet implemented")
}

// FetchNodeDetailJSON fetches node detail via slurmrestd REST API
func (r *RestAPISource) FetchNodeDetailJSON(ctx context.Context, nodeName string) ([]byte, error) {
	return nil, fmt.Errorf("RestAPISource not yet implemented")
}
