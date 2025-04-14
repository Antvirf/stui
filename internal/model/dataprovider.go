package model

import (
	"context"
	"sync"
	"time"
)

// DataProvider is a generic interface for data providers
type DataProvider interface {
	Fetch(ctx context.Context) error
	Data() TableData
	FilteredData(string) TableData
	Subscribe() <-chan struct{}
	LastUpdated() time.Time
	LastError() error
	Close()
}

// BaseProvider contains common provider functionality
type BaseProvider struct {
	mu          sync.RWMutex
	data        TableData
	subscribers []chan struct{}
	lastUpdated time.Time
	lastError   error
	fetchCount  int
	closed      bool
}

// NewBaseProvider creates a new BaseProvider
func NewBaseProvider() BaseProvider {
	return BaseProvider{
		subscribers: make([]chan struct{}, 0),
	}
}

// Fetch should be implemented by concrete providers
func (p *BaseProvider) Fetch(ctx context.Context) error {
	// Concrete providers will override this
	return nil
}

// Data returns a copy of the current data
func (p *BaseProvider) Data() TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a pointer to a copy to prevent external modification
	return *p.data.DeepCopy()
}

// Subscribe returns a channel that receives notifications on data updates
func (p *BaseProvider) Subscribe() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan struct{}, 1)
	p.subscribers = append(p.subscribers, ch)
	return ch
}

// LastUpdated returns the time of the last successful update
func (p *BaseProvider) LastUpdated() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastUpdated
}

// LastError returns the last error that occurred
func (p *BaseProvider) LastError() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastError
}

// Close cleans up all resources
func (p *BaseProvider) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	for _, ch := range p.subscribers {
		close(ch)
	}
	p.subscribers = nil
}

func (p *BaseProvider) FetchCount() int {
	return p.fetchCount
}

// updateData is called by concrete providers when new data is available
func (p *BaseProvider) updateData(data TableData) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data = data
	p.lastUpdated = time.Now()
	p.lastError = nil

	// Notify subscribers
	for _, ch := range p.subscribers {
		select {
		case ch <- struct{}{}:
		default: // Skip if channel is full
		}
	}
}

// updateError is called by concrete providers when an error occurs
func (p *BaseProvider) updateError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lastError = err
}
