package atomic

import (
	"sync"
	"testing"

	"github.com/KiaFarhang/tcp-load-balancer/internal/assert"
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

		assert.Equal(t, counter.Get(), numberOfThreads*operationsPerThread)
	})
	t.Run("Provides thread-safe decrementing", func(t *testing.T) {
		counter := &Counter{count: numberOfThreads * operationsPerThread}

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

		assert.Equal(t, counter.Get(), 0)
	})

}
