package model

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Register types for gob
	gob.Register(&[]config.ColumnConfig{})
	gob.Register(TableData{})
}

// setupTestCache creates a temporary cache for testing
func setupTestCache(t *testing.T) (*SacctCache, string) {
	// Create a temporary directory for test
	tempDir, err := os.MkdirTemp("", "sacct_cache_test")
	require.NoError(t, err)

	// Create cache file path
	filePath := path.Join(tempDir, "sacct_cache.gob")

	// Create a test cache
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	require.NoError(t, err)

	cache := &SacctCache{
		file:   file,
		writer: nil, // will be set after initialization
		reader: nil, // will be set after initialization
	}

	// Set up reader/writer
	cache.writer = bufio.NewWriter(file)
	cache.reader = bufio.NewReader(file)

	return cache, tempDir
}

// createTestTableData creates a TableData for testing
func createTestTableData(t *testing.T, rows int) *TableData {
	headers := []config.ColumnConfig{
		{Name: "JobIDRaw"},
		{Name: "JobID"},
		{Name: "JobName"},
		{Name: "Partition"},
		{Name: "State"},
	}

	tableData := &TableData{
		Headers: &headers,
		Rows:    make([][]string, rows),
	}

	for i := 0; i < rows; i++ {
		tableData.Rows[i] = []string{
			fmt.Sprintf("job_%d", i),      // JobIDRaw
			fmt.Sprintf("%d", i),          // JobID
			fmt.Sprintf("test_job_%d", i), // JobName
			"test_partition",              // Partition
			"RUNNING",                     // State
		}
	}

	return tableData
}

// cleanupTestCache removes the temporary directory and files
func cleanupTestCache(t *testing.T, tempDir string) {
	err := os.RemoveAll(tempDir)
	require.NoError(t, err)
}

// TestBasicCacheOperations tests basic writing and reading operations
func TestBasicCacheOperations(t *testing.T) {
	cache, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Create test data
	testData := createTestTableData(t, 10)
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	// Test writing to cache
	err := cache.WriteToCache(testData, startTime, endTime, true)
	assert.NoError(t, err)

	// Test reading from cache
	readData, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.True(t, cache.IsUsable)

	// Verify data integrity
	assert.Equal(t, len(testData.Rows), len(readData.Rows))
	assert.Equal(t, (*testData.Headers)[0].Name, (*readData.Headers)[0].Name)

	// Check content field is updated
	assert.Equal(t, startTime.Unix(), cache.Content.StartTime.Unix()) // Compare unix timestamps to avoid precision issues
	assert.Equal(t, endTime.Unix(), cache.Content.EndTime.Unix())
}

// TestCorruptedCache tests recovery from corrupted cache file
func TestCorruptedCache(t *testing.T) {
	cache, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Write invalid data to the cache file
	_, err := cache.file.Write([]byte("This is not valid gob data"))
	require.NoError(t, err)
	require.NoError(t, cache.file.Sync())

	// Try to read, should fail but not panic
	_, err = cache.GetFromCache()
	assert.Error(t, err)
	assert.False(t, cache.IsUsable)

	// Now write valid data, should work
	testData := createTestTableData(t, 5)
	err = cache.WriteToCache(testData, time.Now(), time.Now(), true)
	assert.NoError(t, err)

	// Try to read again, should succeed
	readData, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.True(t, cache.IsUsable)
	assert.Equal(t, len(testData.Rows), len(readData.Rows))
}

// TestEmptyCache tests handling of empty cache file
func TestEmptyCache(t *testing.T) {
	cache, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Try to read from empty cache
	data, err := cache.GetFromCache()
	assert.Error(t, err)
	assert.False(t, cache.IsUsable)
	assert.Nil(t, data)

	// Write empty data
	emptyData := createTestTableData(t, 0)
	err = cache.WriteToCache(emptyData, time.Now(), time.Now(), true)
	assert.NoError(t, err)

	// Read empty data
	readData, err := cache.GetFromCache()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(readData.Rows))
}

// TestCacheMerging tests the merging functionality
func TestCacheMerging(t *testing.T) {
	cache, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Create initial data
	initialData := createTestTableData(t, 10)
	startTime := time.Now().Add(-48 * time.Hour)
	midTime := time.Now().Add(-24 * time.Hour)

	// Write initial data
	err := cache.WriteToCache(initialData, startTime, midTime, true)
	assert.NoError(t, err)

	// Create new data (with some overlap to test merging)
	newData := createTestTableData(t, 15) // More rows
	for i := 0; i < 5; i++ {
		// Modify first 5 rows to test replacement
		newData.Rows[i][4] = "COMPLETED" // Change state
	}
	endTime := time.Now()

	// Write new data with merge
	err = cache.WriteToCache(newData, midTime, endTime, false)
	assert.NoError(t, err)

	// Read merged data
	mergedData, err := cache.GetFromCache()
	assert.NoError(t, err)

	// Verify merged data
	assert.GreaterOrEqual(t, len(mergedData.Rows), 15) // Should have at least 15 rows

	// Check if state of first 5 jobs was updated
	foundCompleted := false
	for _, row := range mergedData.Rows {
		if row[0] == "job_0" && row[4] == "COMPLETED" {
			foundCompleted = true
			break
		}
	}
	assert.True(t, foundCompleted, "Modified data should be present in merged result")
}

// TestConcurrentAccess tests thread safety of the cache
func TestConcurrentAccess(t *testing.T) {
	cache, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Write initial data
	initialData := createTestTableData(t, 10)
	err := cache.WriteToCache(initialData, time.Now().Add(-24*time.Hour), time.Now(), true)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	concurrentReaders := 10
	concurrentWriters := 2
	successfulReads := int32(0)
	successfulWrites := int32(0)

	// Start concurrent readers
	for i := 0; i < concurrentReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data, err := cache.GetFromCache()
			if err == nil && len(data.Rows) > 0 {
				atomic.AddInt32(&successfulReads, 1)
			}
		}(i)
	}

	// Start concurrent writers
	for i := 0; i < concurrentWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := createTestTableData(t, 5+id)
			err := cache.WriteToCache(data, time.Now().Add(-12*time.Hour), time.Now(), false)
			if err == nil {
				atomic.AddInt32(&successfulWrites, 1)
			}
		}(i)
	}

	wg.Wait()

	// Verify that operations completed without deadlock
	assert.Greater(t, int(successfulReads), 0)
	assert.Greater(t, int(successfulWrites), 0)
}

// TestCacheFilePermissions tests handling of permission errors
func TestCacheFilePermissions(t *testing.T) {
	// Skip on Windows since permission testing works differently
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir, err := os.MkdirTemp("", "sacct_cache_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	filePath := path.Join(tempDir, "readonly_cache.gob")

	// Create file with read-only permissions
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0400)
	require.NoError(t, err)
	file.Close()

	// Try to open the cache (should fail on write)
	file, err = os.OpenFile(filePath, os.O_RDWR, 0400)
	if err == nil {
		// Some platforms may allow opening but fail on write
		cache := &SacctCache{
			file:   file,
			writer: bufio.NewWriter(file),
			reader: bufio.NewReader(file),
		}

		testData := createTestTableData(t, 5)
		err = cache.WriteToCache(testData, time.Now(), time.Now(), true)
		assert.Error(t, err)
	}
}

// TestCacheReinitialization tests reinitialization from disk
func TestCacheReinitialization(t *testing.T) {
	cache1, tempDir := setupTestCache(t)
	defer cleanupTestCache(t, tempDir)

	// Write data with first instance
	testData := createTestTableData(t, 10)
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	err := cache1.WriteToCache(testData, startTime, endTime, true)
	assert.NoError(t, err)

	// Close first instance
	cache1.file.Close()

	// Create second instance pointing to same file
	filePath := cache1.file.Name()
	file2, err := os.OpenFile(filePath, os.O_RDWR, 0600)
	require.NoError(t, err)

	cache2 := &SacctCache{
		file:   file2,
		writer: bufio.NewWriter(file2),
		reader: bufio.NewReader(file2),
	}

	// Read with second instance
	readData, err := cache2.GetFromCache()
	assert.NoError(t, err)
	assert.True(t, cache2.IsUsable)
	assert.Equal(t, len(testData.Rows), len(readData.Rows))
}
