package loadbalance

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/KiaFarhang/tcp-load-balancer/internal/assert"
)

/*
By default Go runs tests for a package sequentially, so we're fine sharing ports like this
across tests. If we wanted to speed them up and run in parallel we'd have to get fancier (e.g.) dynamically
generate them, but they're running in < .1 seconds as it is.
*/
const (
	loadBalancerPort int = 4444
	upstreamAPort    int = 5555
	upstreamBPort    int = 6666
)

func TestLoadBalancer(t *testing.T) {
	t.Run("Forwards upstream's response to the connected client", func(t *testing.T) {
		upstreamResponse := "Hello World"

		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)
		upstreamAddress := getTCPAddress(t, upstreamAPort)
		loadBalancer, err := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})
		assert.NoError(t, err)

		loadBalancerHandler := func(conn net.Conn) {
			loadBalancer.HandleConnection(context.Background(), conn)
		}

		upstreamHandler := func(conn net.Conn) {
			conn.Write([]byte(upstreamResponse))
			conn.Close()
		}

		loadBalancerServer := newServer(t, loadBalancerAddress, loadBalancerHandler)
		upstreamServer := newServer(t, upstreamAddress, upstreamHandler)

		conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(conn)
		assert.NoError(t, err)
		conn.Close()

		loadBalancerServer.stop()
		upstreamServer.stop()

		assert.Equal(t, string(bytes), upstreamResponse)
	})

	t.Run("Balances between upstreams", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}

		upstreamAResponse := "Hello from upstream A"
		upstreamBResponse := "Hello from upstream B"

		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)
		upstreamAAddress := getTCPAddress(t, upstreamAPort)
		upstreamBAddress := getTCPAddress(t, upstreamBPort)

		loadBalancer, err := NewLoadBalancer([]*net.TCPAddr{upstreamAAddress, upstreamBAddress})
		assert.NoError(t, err)

		loadBalancerHandler := func(conn net.Conn) {
			loadBalancer.HandleConnection(context.Background(), conn)
		}

		upstreamAHandler := func(conn net.Conn) {
			// Hold the connection open; the LB should route a second request to the other upstream
			time.Sleep(3 * time.Second)
			conn.Write([]byte(upstreamAResponse))
			conn.Close()
		}

		upstreamBHandler := func(conn net.Conn) {
			conn.Write([]byte(upstreamBResponse))
			conn.Close()
		}

		loadBalancerServer := newServer(t, loadBalancerAddress, loadBalancerHandler)
		upstreamAServer := newServer(t, upstreamAAddress, upstreamAHandler)
		upstreamBServer := newServer(t, upstreamBAddress, upstreamBHandler)

		var waitGroup sync.WaitGroup

		waitGroup.Add(2)

		var firstConn *net.TCPConn
		var firstConnErr error
		var secondConn *net.TCPConn
		var secondConnErr error

		go func() {
			firstConn, firstConnErr = net.DialTCP("tcp", nil, loadBalancerAddress)
			waitGroup.Done()
		}()

		go func() {
			// Super hacky way to force this connection to come in after the first is waiting, so it'll route to the other upstream
			time.Sleep(1 * time.Second)
			secondConn, secondConnErr = net.DialTCP("tcp", nil, loadBalancerAddress)
			waitGroup.Done()
		}()

		waitGroup.Wait()

		assert.NoError(t, firstConnErr)
		assert.NoError(t, secondConnErr)

		firstResponseBytes, err := io.ReadAll(firstConn)
		assert.NoError(t, err)
		firstConn.Close()

		secondResponseBytes, err := io.ReadAll(secondConn)
		assert.NoError(t, err)
		secondConn.Close()

		loadBalancerServer.stop()
		upstreamAServer.stop()
		upstreamBServer.stop()

		assert.Equal(t, string(firstResponseBytes), upstreamAResponse)
		assert.Equal(t, string(secondResponseBytes), upstreamBResponse)

	})

	t.Run("Returns an error message to client if connection to upstream fails", func(t *testing.T) {
		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)
		upstreamAddress := getTCPAddress(t, upstreamAPort)

		loadBalancer, err := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})
		assert.NoError(t, err)

		loadBalancerHandler := func(conn net.Conn) {
			loadBalancer.HandleConnection(context.Background(), conn)
		}

		loadBalancerServer := newServer(t, loadBalancerAddress, loadBalancerHandler)

		// Deliberately not listening on the upstream address to force a connection error...

		conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(conn)
		assert.NoError(t, err)
		conn.Close()

		loadBalancerServer.stop()

		assert.Equal(t, string(bytes), internalServerErrorMessage)
	})

	t.Run("Returns an error message to client if connection to upstream times out", func(t *testing.T) {
		upstreamResponse := "Hello World"

		loadBalancerAddress := getTCPAddress(t, loadBalancerPort)
		upstreamAddress := getTCPAddress(t, upstreamAPort)

		loadBalancer, err := NewLoadBalancer([]*net.TCPAddr{upstreamAddress})
		assert.NoError(t, err)

		loadBalancerHandler := func(conn net.Conn) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			loadBalancer.HandleConnection(ctx, conn)
		}

		upstreamHandler := func(conn net.Conn) {
			conn.Write([]byte(upstreamResponse))
			conn.Close()
		}

		loadBalancerServer := newServer(t, loadBalancerAddress, loadBalancerHandler)

		upstreamServer := newServer(t, upstreamAddress, upstreamHandler)

		conn, err := net.DialTCP("tcp", nil, loadBalancerAddress)
		assert.NoError(t, err)

		bytes, err := io.ReadAll(conn)
		assert.NoError(t, err)
		conn.Close()

		loadBalancerServer.stop()
		upstreamServer.stop()

		assert.Equal(t, string(bytes), connectionToUpstreamTimedOutMessage)
	})

}

func TestLoadBalancer_findHostWithLeastConnections(t *testing.T) {
	t.Run("Always returns the host with the least connections", func(t *testing.T) {
		host1 := getTCPAddress(t, 1111)

		host2 := getTCPAddress(t, 2222)

		lb, err := NewLoadBalancer([]*net.TCPAddr{host1, host2})
		assert.NoError(t, err)

		for _, host := range lb.hosts {
			if host.address.Port == 1111 {
				host.connectionCount.Increment()
			}
		}

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
}

func getTCPAddress(t *testing.T, port int) *net.TCPAddr {
	t.Helper()
	address, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	assert.NoError(t, err)
	return address
}

/*
A graceful way to spin up + shut down servers for use in tests.
Shamelessly adapted from:
https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
*/
type server struct {
	listener  net.Listener
	handler   func(conn net.Conn)
	quit      chan struct{}
	waitGroup sync.WaitGroup
	t         *testing.T
}

func newServer(t *testing.T, address *net.TCPAddr, handler func(conn net.Conn)) *server {
	server := &server{handler: handler, quit: make(chan struct{}), t: t}

	listener, err := net.ListenTCP("tcp", address)
	assert.NoError(server.t, err)

	server.listener = listener

	server.waitGroup.Add(1)
	go server.serve()

	return server
}

func (s *server) stop() {
	close(s.quit)
	s.listener.Close()
	s.waitGroup.Wait()
}

func (s *server) serve() {
	defer s.waitGroup.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				// Error is fine in this case; we just stopped the server
				return
			default:
				s.t.Errorf("Error accepting connection: %s", err)
			}
		} else {
			s.waitGroup.Add(1)
			go func() {
				s.handler(conn)
				s.waitGroup.Done()
			}()
		}
	}
}
