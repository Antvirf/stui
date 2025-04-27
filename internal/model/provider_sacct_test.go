package model

import (
	"os"
	"testing"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGetSacctData creates a mock for GetSacctData for testing
func mockGetSacctData(data *TableData) func(time.Duration) (*TableData, error) {
	return func(since time.Duration) (*TableData, error) {
		return data, nil
	}
}

// TestSacctProviderFetchToCache tests the FetchToCache functionality
func TestSacctProviderFetchToCache(t *testing.T) {
	// Set up a temporary directory and environment
	tempDir, err := os.MkdirTemp("", "sacct_provider_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original function and restore after test
	originalGetSacctData := GetSacctData
	defer func() { GetSacctData = originalGetSacctData }()

	// Create test data
	testData := createTestTableData(t, 15)

	// Set up mock
	GetSacctData = func(since time.Duration) (*TableData, error) {
		return testData, nil
	}

	// Create provider with test cache
	cache, _ := setupTestCache(t)
	provider := &SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
		cache:        cache,
	}

	// Test FetchToCache
	err = provider.FetchToCache(24 * time.Hour)
	assert.NoError(t, err)

	// Verify cache is updated correctly
	data, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.Equal(t, 15, len(data.Rows))
}

// TestSacctProviderFetchToEmptyCache tests fetching to empty cache
func TestSacctProviderFetchToEmptyCache(t *testing.T) {
	// Set up a temporary directory and environment
	tempDir, err := os.MkdirTemp("", "sacct_provider_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original function and restore after test
	originalGetSacctData := GetSacctData
	defer func() { GetSacctData = originalGetSacctData }()

	// Create test data
	testData := createTestTableData(t, 10)

	// Set up mock
	GetSacctData = mockGetSacctData(testData)

	// Create provider with test cache (empty)
	cache, _ := setupTestCache(t)
	provider := &SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
		cache:        cache,
	}

	// Cache should not be usable initially
	assert.False(t, cache.IsUsable)

	// Fetch to cache
	err = provider.FetchToCache(24 * time.Hour)
	assert.NoError(t, err)

	// After fetching to cache, cache file should now be valid
	// We need to manually call GetFromCache to update the IsUsable flag
	data, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.True(t, cache.IsUsable)
	assert.NotNil(t, data)
}

// TestSacctProviderFilteredData tests the filtering functionality
func TestSacctProviderFilteredData(t *testing.T) {
	// Set up test data with different partitions
	headers := []config.ColumnConfig{
		{Name: "JobIDRaw"},
		{Name: "JobID"},
		{Name: "JobName"},
		{Name: "Partition"},
		{Name: "State"},
	}

	tableData := &TableData{
		Headers: &headers,
		Rows: [][]string{
			{"job_1", "1", "test_job_1", "partition1", "RUNNING"},
			{"job_2", "2", "test_job_2", "partition2", "RUNNING"},
			{"job_3", "3", "test_job_3", "partition1", "COMPLETED"},
			{"job_4", "4", "test_job_4", "partition2", "PENDING"},
		},
	}

	// Create provider with this data
	provider := &SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
	}
	provider.updateData(tableData)

	// Set configuration for testing
	config.SacctViewColumnsPartitionIndex = 3
	config.SacctViewColumnsStateIndex = 4

	// Test filtering by partition
	filtered := provider.FilteredData("partition1")
	assert.Equal(t, 2, len(filtered.Rows))
	assert.Equal(t, "partition1", filtered.Rows[0][3])
	assert.Equal(t, "partition1", filtered.Rows[1][3])

	// Test filtering by state
	config.JobStateCurrentChoice = "RUNNING"
	filtered = provider.FilteredData("")
	assert.Equal(t, 2, len(filtered.Rows))
	assert.Equal(t, "RUNNING", filtered.Rows[0][4])
	assert.Equal(t, "RUNNING", filtered.Rows[1][4])

	// Test combined filtering
	config.JobStateCurrentChoice = "RUNNING"
	filtered = provider.FilteredData("partition1")
	assert.Equal(t, 1, len(filtered.Rows))
	assert.Equal(t, "partition1", filtered.Rows[0][3])
	assert.Equal(t, "RUNNING", filtered.Rows[0][4])

	// Reset state
	config.JobStateCurrentChoice = "(all)"
}

// TestSacctProviderCacheRecovery tests recovery from corrupted cache
func TestSacctProviderCacheRecovery(t *testing.T) {
	// Set up a temporary directory and environment
	tempDir, err := os.MkdirTemp("", "sacct_provider_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original function and restore after test
	originalGetSacctData := GetSacctData
	defer func() { GetSacctData = originalGetSacctData }()

	// Create test data
	testData := createTestTableData(t, 5)

	// Set up mock
	GetSacctData = mockGetSacctData(testData)

	// Create provider with corrupted cache
	cache, _ := setupTestCache(t)
	_, err = cache.file.Write([]byte("corrupted data"))
	require.NoError(t, err)

	provider := &SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
		cache:        cache,
	}

	// Try to fetch - should recover by rewriting cache
	err = provider.FetchToCache(24 * time.Hour)
	assert.NoError(t, err)

	// Try to fetch data
	err = provider.Fetch()
	assert.NoError(t, err)

	// Verify data was recovered
	assert.NotNil(t, provider.Data())
	assert.Equal(t, 5, len(provider.Data().Rows))
}
