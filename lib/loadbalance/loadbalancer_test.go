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

/**
By default Go runs tests for a package sequentially, so we're fine sharing ports like this
across tests. If we wanted to speed them up and run in parallel we'd have to get fancier (e.g.) dynamically
generate them, but they're running in < .1 seconds as it is.
*/
const (
	loadBalancerPort int = 4444
	upstreamAPort    int = 5555
)

func TestLoadBalancer(t *testing.T) {
	t.Run("Forwards upstream's response to the connected client", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		upstreamResponse := "Hello World"

		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)
		loadBalancerListener := getTCPListener(t, loadBalancerAddress)

		upstreamAddress := getTCPAddress(t, upstreamAPort)
		upstreamListener := getTCPListener(t, upstreamAddress)

		loadBalancer := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})

		loadBalancerHandler := func(conn net.Conn) {
			loadBalancer.HandleConnection(context.Background(), conn)
		}

		upstreamHandler := func(conn net.Conn) {
			conn.Write([]byte(upstreamResponse))
			conn.Close()
		}

		runTCPListener(t, loadBalancerListener, loadBalancerHandler)

		runTCPListener(t, upstreamListener, upstreamHandler)

		conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(conn)
		assert.NoError(t, err)

		assert.Equal(t, string(bytes), upstreamResponse)
	})

	t.Run("Returns an error message to client if connection to upstream fails", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)
		loadBalancerListener := getTCPListener(t, loadBalancerAddress)

		upstreamAddress := getTCPAddress(t, upstreamAPort)

		// Deliberately not listening on the upstream address to force a connection error...

		loadBalancer := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})

		loadBalancerHandler := func(conn net.Conn) {
			loadBalancer.HandleConnection(context.Background(), conn)
		}

		runTCPListener(t, loadBalancerListener, loadBalancerHandler)

		conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(conn)
		assert.NoError(t, err)

		assert.Equal(t, string(bytes), internalServerErrorMessage)
	})
}

func TestLoadBalancer_findHostWithLeastConnections(t *testing.T) {
	t.Run("Always returns the host with the least connections", func(t *testing.T) {
		host1 := getTCPAddress(t, 1111)

		host2 := getTCPAddress(t, 2222)

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
		host1 := getTCPAddress(t, 1111)

		host2 := getTCPAddress(t, 2222)

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

func getTCPListener(t *testing.T, address *net.TCPAddr) *net.TCPListener {
	t.Helper()
	listener, err := net.ListenTCP("tcp", address)
	assert.NoError(t, err)

	t.Cleanup(func() {
		listener.Close()
	})

	return listener
}

func runTCPListener(t *testing.T, listener *net.TCPListener, handler func(conn net.Conn)) {
	t.Helper()
	go func() {
		conn, err := listener.Accept()
		assert.NoError(t, err)
		handler(conn)
	}()
}
