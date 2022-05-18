// Package atomic implements a simple threadsafe counter
package atomic

import "sync"

// Counter is a threadsafe integer counter.
type Counter struct {
	count uint
	mu    sync.RWMutex
}

// Increment increments the counter's value.
func (a *Counter) Increment() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count++
}

// Decrement decrements the counter's value.
func (a *Counter) Decrement() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.count--
}

// Get returns the counter's current value.
func (a *Counter) Get() uint {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.count
}
