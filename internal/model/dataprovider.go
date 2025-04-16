package model

import (
	"sync"
	"time"
)

// Internal generic data model
type DataInterface[T any] interface {
	DeepCopy() T
	Length() int
}

// DataProvider is a generic interface for data providers
type DataProvider[T DataInterface[T]] interface {
	Length() int
	Fetch() error
	Data() T
	RunPeriodicRefresh(time.Duration, time.Duration, func())
	FilteredData(string) T
	Subscribe() <-chan struct{}
	LastUpdated() time.Time
	LastError() error
	Close()
}

// BaseProvider[T] contains common provider functionality
type BaseProvider[T DataInterface[T]] struct {
	mu          sync.RWMutex
	data        T
	length      int
	subscribers []chan struct{}
	lastUpdated time.Time
	lastError   error
	fetchCount  int
	closed      bool
}

// NewBaseProvider[T] creates a new BaseProvider[T]
func NewBaseProvider[T DataInterface[T]]() BaseProvider[T] {
	return BaseProvider[T]{
		subscribers: make([]chan struct{}, 0),
	}
}

// Fetch should be implemented by concrete providers
func (p *BaseProvider[T]) Fetch() error {
	return nil
}

// Returns length of data
func (p *BaseProvider[T]) Length() int {
	return p.length
}

// Data returns a copy of the current data
func (p *BaseProvider[T]) Data() T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.DeepCopy()
}

// Subscribe returns a channel that receives notifications on data updates
func (p *BaseProvider[T]) Subscribe() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan struct{}, 1)
	p.subscribers = append(p.subscribers, ch)
	return ch
}

// LastUpdated returns the time of the last successful update
func (p *BaseProvider[T]) LastUpdated() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastUpdated
}

// LastError returns the last error that occurred
func (p *BaseProvider[T]) LastError() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastError
}

// Close cleans up all resources
func (p *BaseProvider[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	for _, ch := range p.subscribers {
		close(ch)
	}
	p.subscribers = nil
}

func (p *BaseProvider[T]) FetchCount() int {
	return p.fetchCount
}

// updateData is called by concrete providers when new data is available
func (p *BaseProvider[T]) updateData(data T) {
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
func (p *BaseProvider[T]) updateError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.lastError = err
}
