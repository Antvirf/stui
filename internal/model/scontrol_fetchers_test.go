package model

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodesWithTimeout(t *testing.T) {
	tests := []struct {
		name            string
		partitionFilter string
		expectedCount   int
		expectedError   bool
		firstNodeName   string
	}{
		{
			name:            "no partition filter",
			partitionFilter: "",
			expectedCount:   888,
			expectedError:   false,
			firstNodeName:   "linux1",
		},
		{
			name:            "with physics partition filter",
			partitionFilter: "physics",
			expectedCount:   100,
			expectedError:   false,
			firstNodeName:   "linux200",
		},
		{
			name:            "with non-existent partition filter",
			partitionFilter: "nonexistent",
			expectedCount:   0,
			expectedError:   false,
			firstNodeName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			config.PartitionFilter = tt.partitionFilter
			data, err := GetNodesWithTimeout(1 * time.Second)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(data.Rows))
			assert.Equal(t, *config.NodeViewColumns, *data.Headers)

			// Verify some sample data
			if tt.expectedCount > 0 {
				assert.Equal(t, tt.firstNodeName, data.Rows[0][0]) // NodeName
			}
		})
	}
}

func TestGetJobs(t *testing.T) {
	config.PartitionFilter = ""
	data, err := GetJobsWithTimeout(1 * time.Second)
	require.NoError(t, err)

	assert.Greater(
		t,
		len(data.Rows),
		0,
		"Expected at least one job to be present, launch a job to run these tests",
	)

	assert.Equal(t, *config.JobViewColumns, *data.Headers)

	firstJobId := data.Rows[0][0]
	details, err := GetJobDetailsWithTimeout(firstJobId, 1*time.Second)
	require.NoError(t, err)
	assert.Contains(t, details, "JobId="+firstJobId)
	assert.Contains(t, details, "JobName=")
}

func TestGetAllPartitionsWithTimeout(t *testing.T) {
	data, err := GetAllPartitionsWithTimeout(1 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, 7, len(data.Rows))

	assert.Equal(t, "general", data.Rows[0][0])
	assert.Equal(t, "chemistry", data.Rows[1][0])
	assert.Equal(t, "physics", data.Rows[2][0])
	assert.Equal(t, "biology", data.Rows[3][0])
	assert.Equal(t, "finance", data.Rows[4][0])
	assert.Equal(t, "mathematics", data.Rows[5][0])
	assert.Equal(t, "unallocated", data.Rows[6][0])
}

func TestGetNodeDetailsWithTimeout(t *testing.T) {
	details, err := GetNodeDetailsWithTimeout("linux1", 1*time.Second)
	require.NoError(t, err)
	assert.Contains(t, details, "NodeName=linux1")
	assert.Contains(t, details, "CPUTot=64")
}

func TestGetSchedulerInfoWithTimeout(t *testing.T) {
	host, _ := GetSchedulerInfoWithTimeout(1 * time.Second)
	testRunnerHostName, _ := os.Hostname()
	assert.Equal(t, fmt.Sprintf("%s(localhost)", testRunnerHostName), host)
}

func TestGetSdiagWithTimeout(t *testing.T) {
	output, err := GetSdiagWithTimeout(1 * time.Second)
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
