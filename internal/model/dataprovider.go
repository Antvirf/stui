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
	FilteredData(string) T
	LastUpdated() time.Time
	LastError() error
}

// BaseProvider[T] contains common provider functionality
type BaseProvider[T DataInterface[T]] struct {
	mu          sync.RWMutex
	data        T
	length      int
	lastUpdated time.Time
	lastError   error
	fetchCount  int
}

// NewBaseProvider[T] creates a new BaseProvider[T]
func NewBaseProvider[T DataInterface[T]]() BaseProvider[T] {
	return BaseProvider[T]{}
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

func (p *BaseProvider[T]) FetchCount() int {
	return p.fetchCount
}

// updateData is called by concrete providers when new data is available
func (p *BaseProvider[T]) updateData(data T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.data = data
	p.fetchCount += 1
	p.lastUpdated = time.Now()
	p.lastError = nil
	p.length = p.data.Length()
}

// updateError is called by concrete providers when an error occurs
func (p *BaseProvider[T]) updateError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.length = 0
	p.lastError = err
}
