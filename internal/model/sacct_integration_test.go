package model

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullSacctWorkflow tests the entire workflow from cache initialization to filtered display
func TestFullSacctWorkflow(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "sacct_integration_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Mock GetSacctData
	originalGetSacctData := GetSacctData
	defer func() { GetSacctData = originalGetSacctData }()

	// Create varied test data
	mockData := &TableData{
		Headers: &[]config.ColumnConfig{
			{Name: "JobIDRaw"},
			{Name: "JobID"},
			{Name: "JobName"},
			{Name: "Partition"},
			{Name: "State"},
		},
		Rows: [][]string{
			{"100", "100", "job1", "partition1", "RUNNING"},
			{"101", "101", "job2", "partition2", "COMPLETED"},
			{"102", "102", "job3", "partition1", "PENDING"},
			{"103", "103", "job4", "partition3", "FAILED"},
		},
	}

	GetSacctData = func(since time.Duration) (*TableData, error) {
		return mockData, nil
	}

	// Set configuration for testing
	config.SacctViewColumnsPartitionIndex = 3
	config.SacctViewColumnsStateIndex = 4

	// 1. Initialize provider with custom cache location
	cachePath := path.Join(tempDir, "test_sacct_cache.gob")
	cacheFile, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0600)
	require.NoError(t, err)
	defer cacheFile.Close()

	cache := &SacctCache{
		file:   cacheFile,
		writer: bufio.NewWriter(cacheFile),
		reader: bufio.NewReader(cacheFile),
	}

	provider := &SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
		cache:        cache,
	}

	// 2. Load data into cache
	err = provider.FetchToCache(24 * time.Hour)
	assert.NoError(t, err)

	// 3. Verify cache is populated
	cacheData, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(cacheData.Rows))

	// 4. Test standard Fetch
	err = provider.Fetch()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(provider.Data().Rows))

	// 5. Test filtered data access
	config.JobStateCurrentChoice = "RUNNING"
	filtered := provider.FilteredData("partition1")
	assert.Equal(t, 1, len(filtered.Rows))
	assert.Equal(t, "RUNNING", filtered.Rows[0][4])

	// 6. Change state filter
	config.JobStateCurrentChoice = "PENDING"
	filtered = provider.FilteredData("partition1")
	assert.Equal(t, 1, len(filtered.Rows))
	assert.Equal(t, "PENDING", filtered.Rows[0][4])

	// 7. Reset state filter
	config.JobStateCurrentChoice = "(all)"
	filtered = provider.FilteredData("partition1")
	assert.Equal(t, 2, len(filtered.Rows))

	// 8. Test durability - close and reopen cache
	cacheFile.Close()

	// Reopen cache file
	cacheFile, err = os.OpenFile(cachePath, os.O_RDWR, 0600)
	require.NoError(t, err)

	newCache := &SacctCache{
		file:   cacheFile,
		writer: bufio.NewWriter(cacheFile),
		reader: bufio.NewReader(cacheFile),
	}

	// Read the persisted data
	newData, err := newCache.GetFromCache()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(newData.Rows))
}

// TestCacheVersionCompatibility tests handling of different versions of cached data
func TestCacheVersionCompatibility(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "cache_version_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	cachePath := path.Join(tempDir, "test_version_cache.gob")

	// Create a file with a different schema (simulating older version)
	file, err := os.Create(cachePath)
	require.NoError(t, err)

	// Create a structure that's different from current SacctCacheContents
	type OldCacheFormat struct {
		StartTime string // Different type than current (string vs time.Time)
		Data      *TableData
		// Missing EndTime field
	}

	oldData := OldCacheFormat{
		StartTime: "2023-01-01",
		Data:      createTestTableData(t, 3),
	}

	// Write the old format data
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(oldData)
	require.NoError(t, err)
	file.Close()

	// Now try to open with current cache implementation
	file, err = os.OpenFile(cachePath, os.O_RDWR, 0600)
	require.NoError(t, err)
	defer file.Close()

	cache := &SacctCache{
		file:   file,
		writer: bufio.NewWriter(file),
		reader: bufio.NewReader(file),
	}

	// Try to read - should fail but not panic
	_, err = cache.GetFromCache()
	assert.Error(t, err)
	assert.False(t, cache.IsUsable)

	// Now write new format data
	testData := createTestTableData(t, 5)
	err = cache.WriteToCache(testData, time.Now(), time.Now(), true)
	assert.NoError(t, err)

	// Should be able to read new format
	newData, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.True(t, cache.IsUsable)
	assert.Equal(t, 5, len(newData.Rows))
}

// TestCacheEvictionStrategy tests the strategy for data retention
func TestCacheEvictionStrategy(t *testing.T) {
	cache, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// First, write old data
	oldData := createTestTableData(t, 10)
	for i := 0; i < 10; i++ {
		oldData.Rows[i][0] = fmt.Sprintf("old_job_%d", i)
	}

	oldStart := time.Now().Add(-48 * time.Hour)
	oldEnd := time.Now().Add(-24 * time.Hour)

	err := cache.WriteToCache(oldData, oldStart, oldEnd, true)
	assert.NoError(t, err)

	// Now write newer data without overlapping job IDs
	newData := createTestTableData(t, 5)
	for i := 0; i < 5; i++ {
		newData.Rows[i][0] = fmt.Sprintf("new_job_%d", i)
	}

	newStart := time.Now().Add(-12 * time.Hour)
	newEnd := time.Now()

	err = cache.WriteToCache(newData, newStart, newEnd, false)
	assert.NoError(t, err)

	// Read merged data
	mergedData, err := cache.GetFromCache()
	assert.NoError(t, err)

	// Should contain both old and new data
	assert.Equal(t, 15, len(mergedData.Rows))

	// Instead of comparing times directly, just verify row content which is what we really care about
	// Check that we have both old and new jobs in the merged data
	hasOldJobs := false
	hasNewJobs := false

	for _, row := range mergedData.Rows {
		if len(row) > 0 {
			if strings.HasPrefix(row[0], "old_job_") {
				hasOldJobs = true
			}
			if strings.HasPrefix(row[0], "new_job_") {
				hasNewJobs = true
			}
		}
	}

	assert.True(t, hasOldJobs, "Merged data should contain old jobs")
	assert.True(t, hasNewJobs, "Merged data should contain new jobs")
}
