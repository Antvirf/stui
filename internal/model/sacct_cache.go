package model

import (
	"bufio"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/antvirf/stui/internal/config"
)

func init() {
	// Must register types before using them with gob
	gob.Register(&[]config.ColumnConfig{})
	gob.Register(TableData{})
}

type SacctCache struct {
	mu       sync.RWMutex
	file     *os.File
	writer   *bufio.Writer
	reader   *bufio.Reader
	init     sync.Once
	IsUsable bool
	Content  SacctCacheContents
}

type SacctCacheContents struct {
	StartTime time.Time
	EndTime   time.Time
	Data      *TableData
}

func NewSacctCache() (*SacctCache, error) {
	s := SacctCache{}

	var initErr error
	s.init.Do(func() {

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return
		}

		filePath := path.Join(
			homeDir,
			".cache",
			fmt.Sprintf(
				"stui_sacct_cache_%s.gob",
				hex.EncodeToString([]byte(config.ClusterName)),
			),
		)

		if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
			initErr = err
			return
		}

		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			initErr = err
			return
		}

		s.file = file
		s.writer = bufio.NewWriter(file)
		s.reader = bufio.NewReader(file)
	})

	// Try to read from the cache to see if it's initialized and not corrupted
	// as well as load up cache content.
	_, err := s.GetFromCache()
	s.IsUsable = (err == nil)

	return &s, initErr
}

func (c *SacctCache) WriteToCache(data *TableData, start, end time.Time, rewriteEntireCache bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := SacctCacheContents{
		StartTime: start,
		EndTime:   end,
		Data:      data,
	}

	if rewriteEntireCache || !c.IsUsable {
		// Clear the cache by truncating the file
		if err := c.file.Truncate(0); err != nil {
			return err
		}
		if _, err := c.file.Seek(0, 0); err != nil {
			return err
		}

		// Create a fresh encoder
		encoder := gob.NewEncoder(c.writer)
		if err := encoder.Encode(entry); err != nil {
			return err
		}

		if err := c.writer.Flush(); err != nil {
			return err
		}

		// Make sure data is persisted to disk
		if err := c.file.Sync(); err != nil {
			return err
		}

		// Update the content and mark cache as usable
		c.Content = entry
		c.IsUsable = true

		// Reset the reader to reflect new file content
		c.reader = bufio.NewReader(c.file)

		return nil
	}

	// Merge new data with existing cache
	existingData, err := c.GetFromCache()
	if err != nil {
		return err
	}

	mergedData := mergeTableData(existingData, data)

	mergedEntry := SacctCacheContents{
		StartTime: start,
		EndTime:   end,
		Data:      mergedData,
	}

	// Update the object
	c.Content = mergedEntry
	c.IsUsable = true

	// Update cache file
	if err := c.file.Truncate(0); err != nil {
		return err
	}
	if _, err := c.file.Seek(0, 0); err != nil {
		return err
	}

	if err := gob.NewEncoder(c.writer).Encode(mergedEntry); err != nil {
		return err
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}

	// Reset the reader to reflect new file content
	c.reader = bufio.NewReader(c.file)

	return nil
}

func mergeTableData(oldData, newData *TableData) *TableData {
	// Construct headers based on hardcoded config
	var headers []config.ColumnConfig
	columnConfig := strings.Split(SACCT_COLUMNS, ",")
	for _, key := range columnConfig {
		headers = append(headers, config.ColumnConfig{Name: key})
	}

	merged := &TableData{
		Headers: &headers,
		Rows:    make([][]string, 0, len(oldData.Rows)+len(newData.Rows)),
	}

	// Create a map for quick lookup of new data by JobIDRaw
	newDataMap := make(map[string][]string)
	for _, row := range newData.Rows {
		if len(row) > 0 {
			newDataMap[row[0]] = row // Assuming the first column is JobIDRaw
		}
	}

	replacedLines := 0
	keptLines := 0
	addedLines := 0

	// Add or replace rows from old data
	for _, row := range oldData.Rows {
		if len(row) > 0 {
			if newRow, exists := newDataMap[row[0]]; exists {
				merged.Rows = append(merged.Rows, newRow)
				delete(newDataMap, row[0]) // Remove to avoid duplicates
				replacedLines++
			} else {
				merged.Rows = append(merged.Rows, row)
				keptLines++
			}
		}
	}

	// Add remaining new rows
	for _, row := range newDataMap {
		merged.Rows = append(merged.Rows, row)
		addedLines++
	}

	// Sort rows by the first column (JobIDRaw)
	sort.Slice(merged.Rows, func(i, j int) bool {
		return merged.Rows[i][0] > merged.Rows[j][0]
	})

	// Total rows
	return merged
}

func (c *SacctCache) GetFromCache() (*TableData, error) {
	// Reset to beginning of file
	if _, err := c.file.Seek(0, 0); err != nil {
		c.IsUsable = false
		return nil, err
	}

	// Refresh the reader
	c.reader = bufio.NewReader(c.file)

	var entry SacctCacheContents
	if err := gob.NewDecoder(c.reader).Decode(&entry); err != nil {
		c.IsUsable = false
		return nil, err
	}

	// If the cache is empty or data is invalid, return an empty TableData
	if entry.Data == nil || len(entry.Data.Rows) == 0 {
		c.IsUsable = false
		return &TableData{
			Headers: &[]config.ColumnConfig{},
			Rows:    [][]string{},
		}, nil
	}

	c.Content = entry
	c.IsUsable = true

	return entry.Data, nil
}
