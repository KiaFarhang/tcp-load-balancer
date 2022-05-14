package loadbalance

import (
	"net"
	"sync"
	"testing"

	"github.com/KiaFarhang/tcp-load-balancer/lib/assert"
)

func TestLoadBalancer_findHostWithLeastConnections(t *testing.T) {
	t.Run("Always returns the host with the least connections", func(t *testing.T) {
		host1, err := net.ResolveTCPAddr("tcp", ":1111")
		assert.NoError(t, err)

		host2, err := net.ResolveTCPAddr("tcp", ":2222")
		assert.NoError(t, err)

		lb := NewLoadBalancer([]*net.TCPAddr{host1, host2})

		lb.hosts[0].connectionCount.Increment()

		var wg sync.WaitGroup

		wg.Add(100)

		// No matter how many times we call it, we should always get host2 - it has no conns
		for i := 0; i < 100; i++ {
			go func() {
				h := lb.findHostWithLeastConnections()
				assert.Equal(t, h.address, host2)
				wg.Done()
			}()
		}

		wg.Wait()
	})
	t.Run("Defaults to lower-index host in case of a tie", func(t *testing.T) {
		host1, err := net.ResolveTCPAddr("tcp", ":1111")
		assert.NoError(t, err)

		host2, err := net.ResolveTCPAddr("tcp", ":2222")
		assert.NoError(t, err)

		lb := NewLoadBalancer([]*net.TCPAddr{host1, host2})

		var wg sync.WaitGroup

		wg.Add(100)

		for i := 0; i < 100; i++ {
			go func() {
				h := lb.findHostWithLeastConnections()
				assert.Equal(t, h.address, host1)
				wg.Done()
			}()
		}

		wg.Wait()
	})
}
