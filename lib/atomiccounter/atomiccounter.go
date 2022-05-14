// Package atomiccounter implements a simple threadsafe counter
package atomiccounter

import "sync"

// AtomicCounter is a threadsafe integer counter
type AtomicCounter struct {
	count int
	mu    sync.RWMutex
}

// Increment increments the counter's value
func (a *AtomicCounter) Increment() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count++
}

// Decrement decrements the counter's value
func (a *AtomicCounter) Decrement() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count--
}

// Get returns the counter's current value
func (a *AtomicCounter) Get() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.count
}
