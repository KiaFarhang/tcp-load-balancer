// Package atomic implements a simple threadsafe counter
package atomic

import (
	"sync/atomic"
)

// Counter is a threadsafe integer counter.
type Counter struct {
	count uint64
}

// Increment increments the counter's value.
func (a *Counter) Increment() {
	atomic.AddUint64(&a.count, 1)
}

// Decrement decrements the counter's value.
func (a *Counter) Decrement() {
	atomic.AddUint64(&a.count, ^uint64(0))
}

// Get returns the counter's current value.
func (a *Counter) Get() uint64 {
	return atomic.LoadUint64(&a.count)
}
