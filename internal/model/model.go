package model

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/antvirf/stui/internal/config"
)

// TableData represents the data returned by the model package, ready for display.
// Rows contain typed CellValue instances that maintain column-level type consistency.
type TableData struct {
	Headers             *[]config.ColumnConfig
	Rows                [][]CellValue // List of typed cell values
	RowsAsSingleStrings []string      // List of strings - used for searching
}

func EmptyTableData() *TableData {
	return &TableData{
		Headers:             &[]config.ColumnConfig{},
		Rows:                [][]CellValue{},
		RowsAsSingleStrings: []string{},
	}
}

func EmptyTableDataWithColumns(columns *[]config.ColumnConfig) *TableData {
	return &TableData{
		Headers:             columns,
		Rows:                [][]CellValue{},
		RowsAsSingleStrings: []string{},
	}
}

func convertRowsToRowsAsSingleStrings(rows [][]CellValue, headers *[]config.ColumnConfig) []string {
	// Identify NodeList column indices for expansion
	var nodeListIndices []int
	if headers != nil {
		for i, col := range *headers {
			if col.DisplayName == "NodeList" {
				nodeListIndices = append(nodeListIndices, i)
			}
		}
	}

	rowsAsStrings := []string{}
	for _, row := range rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = cell.Display()
		}
		rowStr := strings.Join(cells, "")

		// Append expanded node lists for searchability
		for _, idx := range nodeListIndices {
			if idx < len(row) {
				display := row[idx].Display()
				if strings.Contains(display, "[") {
					expanded := ExpandNodeList(display)
					rowStr += expanded
				}
			}
		}

		rowsAsStrings = append(rowsAsStrings, rowStr)
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

	rowsCopy := make([][]CellValue, len(t.Rows))
	for i, row := range t.Rows {
		rowCopy := make([]CellValue, len(row))
		// Shallow copy is safe because all CellValue implementations are immutable
		// (no methods modify internal state after creation)
		copy(rowCopy, row)
		rowsCopy[i] = rowCopy
	}

	return &TableData{
		Headers:             copiedHeaders,
		Rows:                rowsCopy,
		RowsAsSingleStrings: convertRowsToRowsAsSingleStrings(rowsCopy, copiedHeaders),
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

// filterSplitRe splits state/partition fields on "+" or "," delimiters.
var filterSplitRe = regexp.MustCompile("[,+]")

// Applies a list of given filters to the data
func (t *TableData) ApplyFilters(filters map[int]string) *TableData {
	rows := make([][]CellValue, 0, len(t.Rows))
	rowsAsSingleStrings := make([]string, 0, len(t.Rows))
rowLoop:
	for i, row := range t.Rows {
		for filterKey, filterValue := range filters {
			if filterValue != config.ALL_CATEGORIES_OPTION {

				// The filters we have are either by state (+-separated) or by partition (comma-separated). We split by both.
				// Use Display() to get string representation for filtering
				valuesInRowField := filterSplitRe.Split(row[filterKey].Display(), -1)
				if !slices.Contains(valuesInRowField, filterValue) {
					continue rowLoop
				}
			}
		}
		rows = append(rows, row)
		rowsAsSingleStrings = append(rowsAsSingleStrings, t.RowsAsSingleStrings[i])
	}

	return &TableData{
		Headers:             t.Headers,
		Rows:                rows,
		RowsAsSingleStrings: rowsAsSingleStrings,
	}
}

func (td *TableData) rowToMap(row []CellValue) map[string]string {
	data := make(map[string]string)
	for i, header := range *td.Headers {
		if i < len(row) {
			// Convert header name to Go-template-friendly format
			key := strings.ReplaceAll(header.RawName, " ", "_")
			// Use Display() for template compatibility
			data[key] = row[i].Display()
		}
	}
	return data
}

func (td *TableData) GetRowAsMapById(idString string) (map[string]string, error) {
	for _, row := range td.Rows {
		if len(row) > 0 && row[0].Display() == idString {
			return td.rowToMap(row), nil
		}
	}
	return nil, errors.New("not found")
}
