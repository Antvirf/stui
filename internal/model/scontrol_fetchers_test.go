package model

import (
	"testing"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodesProvider(t *testing.T) {
	provider := NewNodesProvider()

	tests := []struct {
		name            string
		partitionFilter string
		expectedCount   int
	}{
		{
			name:            "no partition filter",
			partitionFilter: config.ALL_CATEGORIES_OPTION,
			expectedCount:   8888,
		},
		{
			name:            "with physics partition filter",
			partitionFilter: "physics",
			expectedCount:   100,
		},
		{
			name:            "with non-existent partition filter",
			partitionFilter: "nonexistent",
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.PartitionFilter = tt.partitionFilter
			config.NodeStateCurrentChoice = config.ALL_CATEGORIES_OPTION
			data := provider.FilteredData()
			assert.Equal(t, tt.expectedCount, len(data.Rows))
			assert.Equal(t, *config.NodeViewColumns, *data.Headers)

			if tt.expectedCount > 0 {
				assert.NotEmpty(t, data.Rows[0][0].Display()) // NodeName
			}
		})
	}
}

func TestJobsProvider(t *testing.T) {
	provider := NewJobsProvider()

	config.PartitionFilter = config.ALL_CATEGORIES_OPTION
	data := provider.FilteredData()
	assert.Greater(t, len(data.Rows), 0, "Expected at least one job")
	assert.Equal(t, *config.JobViewColumns, *data.Headers)

	firstJobId := data.Rows[0][0].Display()
	details, err := GetJobDetailsWithTimeout(firstJobId, 1*time.Second)
	require.NoError(t, err)
	assert.Contains(t, details, "JobId="+firstJobId)
	assert.Contains(t, details, "JobName=")
}

func TestPartitionsProvider(t *testing.T) {
	provider := NewPartitionsProvider()

	data := provider.FilteredData()
	assert.Greater(t, len(data.Rows), 0, "Should have at least one partition")
	// Just check that we have some partition data - order may vary
	assert.NotEmpty(t, data.Rows[0][0].Display())
	// Check that rows are CellValue types
	assert.Implements(t, (*CellValue)(nil), data.Rows[0][0])
}

func TestGetNodeDetailsWithTimeout(t *testing.T) {
	details, err := GetNodeDetailsWithTimeout("linux1", 1*time.Second)
	require.NoError(t, err)
	assert.Contains(t, details, "NodeName=linux1")
	assert.Contains(t, details, "CPUTot=2") // Actual value from test cluster
}

func TestGetSdiagWithTimeout(t *testing.T) {
	output, err := getSdiagWithTimeout(1 * time.Second)
	require.NoError(t, err)
	assert.Contains(t, output, "Server thread count")
	assert.Contains(t, output, "Jobs submitted")
	assert.Contains(t, output, "Main schedule statistics")
	assert.Contains(t, output, "Backfilling stats")
	assert.Contains(t, output, "Remote Procedure Call statistics by message type")
}

func init() {
	config.ComputeConfigurations()
}
