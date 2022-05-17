package loadbalance

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/KiaFarhang/tcp-load-balancer/internal/assert"
)

const (
	loadBalancerPort int = 4444
	upstreamAPort    int = 5555
)

func TestLoadBalancer(t *testing.T) {
	t.Run("Forwards upstream's request to the connected client", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)

		loadBalancerListener, err := net.ListenTCP("tcp", loadBalancerAddress)
		assert.NoError(t, err)

		defer loadBalancerListener.Close()

		upstreamAddress := getTCPAddress(t, upstreamAPort)

		upstreamListener, err := net.ListenTCP("tcp", upstreamAddress)

		assert.NoError(t, err)

		defer upstreamListener.Close()

		loadBalancer := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})

		loadBalancerHandler := func(conn net.Conn) {
			loadBalancer.HandleConnection(context.Background(), conn)
		}

		upstreamHandler := func(conn net.Conn) {
			conn.Write([]byte("Hello World"))
			conn.Close()
		}

		go func() {
			conn, err := loadBalancerListener.Accept()
			assert.NoError(t, err)
			loadBalancerHandler(conn)
		}()

		go func() {
			conn, err := upstreamListener.Accept()
			assert.NoError(t, err)
			upstreamHandler(conn)
		}()

		conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(conn)
		assert.NoError(t, err)

		assert.Equal(t, string(bytes), "Hello World")
	})
}

func TestLoadBalancer_ErrorConnectingToUpstream(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	loadBalancerAddress, err := net.ResolveTCPAddr("tcp", ":6666")
	assert.NoError(t, err)

	loadBalancerListener, err := net.ListenTCP("tcp", loadBalancerAddress)

	assert.NoError(t, err)

	defer loadBalancerListener.Close()

	upstreamAddress, err := net.ResolveTCPAddr("tcp", ":7777")
	assert.NoError(t, err)

	// Deliberately not listening on the upstream address to force a connection error...

	loadBalancer := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})

	loadBalancerHandler := func(conn net.Conn) {
		loadBalancer.HandleConnection(context.Background(), conn)
	}

	go func() {
		conn, err := loadBalancerListener.Accept()
		assert.NoError(t, err)
		loadBalancerHandler(conn)
	}()

	conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
	assert.NoError(t, err)

	bytes, err := io.ReadAll(conn)
	assert.NoError(t, err)

	assert.Equal(t, string(bytes), internalServerErrorMessage)
}

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

func getTCPAddress(t *testing.T, port int) *net.TCPAddr {
	t.Helper()
	address, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	assert.NoError(t, err)
	return address
}
