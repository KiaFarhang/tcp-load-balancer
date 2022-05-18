package atomic

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	numberOfThreads     int = 50
	operationsPerThread int = 1000
)

func TestAtomicCounter(t *testing.T) {
	t.Run("Provides thread-safe incrementing", func(t *testing.T) {
		counter := &Counter{}

		var waitGroup sync.WaitGroup

		for i := 0; i < numberOfThreads; i++ {
			waitGroup.Add(1)

			go func() {
				for c := 0; c < operationsPerThread; c++ {
					counter.Increment()
				}
				waitGroup.Done()
			}()
		}

		waitGroup.Wait()

		assert.Equal(t, uint64(numberOfThreads*operationsPerThread), counter.Get())
	})
	t.Run("Provides thread-safe decrementing", func(t *testing.T) {
		counter := &Counter{count: uint64(numberOfThreads * operationsPerThread)}

		var waitGroup sync.WaitGroup

		for i := 0; i < numberOfThreads; i++ {
			waitGroup.Add(1)

			go func() {
				for c := 0; c < operationsPerThread; c++ {
					counter.Decrement()
				}
				waitGroup.Done()
			}()
		}

		waitGroup.Wait()

		assert.Equal(t, uint64(0), counter.Get())
	})

}

func BenchmarkAutomicCounter(b *testing.B) {
	counter := &Counter{}

	var waitGroup sync.WaitGroup

	for i := 0; i < b.N; i++ {
		waitGroup.Add(1)

		go func() {
			for c := 0; c < operationsPerThread; c++ {
				counter.Increment()
			}
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	assert.Equal(b, uint64(b.N*operationsPerThread), counter.Get())
}
