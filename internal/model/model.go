package model

import (
	"sync"

	"github.com/antvirf/stui/internal/config"
)

// TableData represents the data returned by the model package, ready for display.
type TableData struct {
	Headers *[]config.ColumnConfig
	Rows    [][]string
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
