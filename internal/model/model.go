package model

import (
	"errors"
	"strings"
	"sync"

	"github.com/antvirf/stui/internal/config"
)

// TableData represents the data returned by the model package, ready for display.
type TableData struct {
	Headers             *[]config.ColumnConfig
	Rows                [][]string // List of lists
	RowsAsSingleStrings []string   // List of strings - used for searching
}

func EmptyTableData() *TableData {
	return &TableData{
		Headers:             &[]config.ColumnConfig{},
		Rows:                [][]string{},
		RowsAsSingleStrings: []string{},
	}
}

func convertRowsToRowsAsSingleStrings(rows [][]string) []string {
	rowsAsStrings := []string{}
	for _, row := range rows {
		rowsAsStrings = append(rowsAsStrings, strings.Join(row, ""))
	}
	return rowsAsStrings
}

// DeepCopy creates a deep copy of the TableData struct.
func (t *TableData) DeepCopy() *TableData {
	var copiedHeaders *[]config.ColumnConfig
	if t.Headers != nil {
		headersCopy := make([]config.ColumnConfig, len(*t.Headers))
		copy(headersCopy, *t.Headers)
		copiedHeaders = &headersCopy
	}

	rowsCopy := make([][]string, len(t.Rows))
	for i, row := range t.Rows {
		rowCopy := make([]string, len(row))
		copy(rowCopy, row)
		rowsCopy[i] = rowCopy
	}

	return &TableData{
		Headers:             copiedHeaders,
		Rows:                rowsCopy,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rowsCopy),
	}
}

func (t *TableData) Length() int {
	return len(t.Rows)
}

// TextData is an internal data structure used by providers to store text
type TextData struct {
	Data string
}

func (t *TextData) DeepCopy() *TextData {
	d := TextData{Data: strings.Clone(t.Data)}
	return &d
}

func (t *TextData) Length() int {
	return len(t.Data)
}

var (
	FetchCounter threadSafeCounter // Counter for total number of data fetches
)

type threadSafeCounter struct {
	Count int
	mu    sync.Mutex
}

func (c *threadSafeCounter) increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Count++
}

func init() {
	FetchCounter = threadSafeCounter{Count: 0}

	// We increment count here, because we do a call during
	// config initialization.
	FetchCounter.increment()
}

// Applies a list of given filters to the data
func (t *TableData) ApplyFilters(filters map[int]string) *TableData {
	data := t.DeepCopy()

	var rows [][]string
rowLoop:
	for _, row := range data.Rows {
		for filterKey, filterValue := range filters {
			if filterValue != config.ALL_CATEGORIES_OPTION {
				if !strings.Contains(row[filterKey], filterValue) {
					continue rowLoop
				}
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers:             data.Headers,
		Rows:                rows,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rows),
	}
}

func (td *TableData) rowToMap(row []string) map[string]string {
	data := make(map[string]string)
	for i, header := range *td.Headers {
		if i < len(row) {
			// Convert header name to Go-template-friendly format
			key := strings.ReplaceAll(header.RawName, " ", "_")
			data[key] = row[i]
		}
	}
	return data
}

func (td *TableData) GetRowAsMapById(idString string) (map[string]string, error) {
	for _, row := range td.Rows {
		if len(row) > 0 && row[0] == idString {
			return td.rowToMap(row), nil
		}
	}
	return nil, errors.New("not found")
}
