package model

import (
	"strings"
	"sync"

	"github.com/antvirf/stui/internal/config"
)

// TableData represents the data returned by the model package, ready for display.
type TableData struct {
	Headers *[]config.ColumnConfig
	Rows    [][]string
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
		Headers: copiedHeaders,
		Rows:    rowsCopy,
	}
}

// TextData is an internal data structure used by providers to store text
type TextData struct {
	Data string
}

func (t *TextData) DeepCopy() *TextData {
	d := TextData{Data: strings.Clone(t.Data)}
	return &d
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
}
