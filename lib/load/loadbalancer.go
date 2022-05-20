// Package load implements a least-connections load balancer.
package load

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxConnectionTimeout time.Duration = 3 * time.Second
)

// host is an individual server that can take traffic from the load balancer.
type host struct {
	address         *net.TCPAddr
	connectionCount uint64
}

/*
Balancer is a least-connections load balancer. It keeps track of the number
of connections to a group of hosts, and routes a request to whichever host has the fewest
at the time the request is processed.

If two hosts have the same number of connections, the Balancer will select a host at random.
*/
type Balancer struct {
	hosts  []*host
	dialer *net.Dialer
	mu     sync.Mutex
}

/*
NewLoadBalancer constructs a new least-connections load balancer to route
requests to the slice of TCP addresses provided. Clients should construct a separate
LoadBalancer for each upstream application they wish to load balance.

An error is returned in any of the following scenarios:

- The slice of addresses passed is empty
- The slice of addresses passed is nil
- The slice of addresses contains only nil entries

Duplicate addresses are ignored. If two addresses passed share the same IP, zone and port they
will be treated as a single host when performing load balancing.
*/
func NewLoadBalancer(addresses []*net.TCPAddr) (*Balancer, error) {
	validateAddresses, err := validateAndRemoveDuplicateAddresses(addresses)
	if err != nil {
		return nil, err
	}

	hosts := make([]*host, 0, len(validateAddresses))
	for _, address := range validateAddresses {
		host := &host{address: address, connectionCount: 0}
		hosts = append(hosts, host)
	}

	// When both a context and a dialer have a timeout the shorter value
	// is respected; this protects us from clients passing in a no-timeout context
	// and our dial deadlocking when we can't connect to the upstream.
	return &Balancer{hosts: hosts, dialer: &net.Dialer{Timeout: maxConnectionTimeout}, mu: sync.Mutex{}}, nil
}

/*
HandleConnection takes a TCP connection, finds a suitable host to handle it,
then connects to the host and streams data between the connection and host until
both sides of the connection are closed.

If the connection to the host fails, HandleConnection will close the incoming net.Conn.
It will also close the net.Conn if the communication with the host succeeds; callers of
HandleConnection do not need to close the net.Conn they pass in.
*/
func (b *Balancer) HandleConnection(ctx context.Context, conn net.Conn) {
	host := b.findHostWithLeastConnections()

	atomic.AddUint64(&host.connectionCount, 1)
	defer atomic.AddUint64(&host.connectionCount, ^uint64(0))

	connectionToHost, err := b.dialer.DialContext(ctx, "tcp", host.address.String())

	if err != nil {
		// TODO: in a real system we'd probably want to inject the logger via constructor
		log.Printf("Error dialing host IP %s: %s", host.address.IP.String(), err.Error())
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("Error closing connection after dial failure. Dial error: %s, connection close error: %s", err.Error(), closeErr.Error())
		}
		return
	}

	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	go func() {
		defer conn.Close()
		_, err := io.Copy(conn, connectionToHost)
		if err != nil {
			log.Printf("Error copying from host to client: %s", err.Error())
		}
		waitGroup.Done()
	}()

	go func() {
		defer connectionToHost.Close()
		_, err := io.Copy(connectionToHost, conn)
		if err != nil {
			log.Printf("Error copying from client to host: %s", err.Error())
		}
		waitGroup.Done()
	}()

	waitGroup.Wait()
}

func (b *Balancer) findHostWithLeastConnections() *host {
	host := b.hosts[0]

	b.mu.Lock()
	defer b.mu.Unlock()
	for _, h := range b.hosts[1:] {
		hConnectionCount := atomic.LoadUint64(&h.connectionCount)
		hostConnectionCount := atomic.LoadUint64(&host.connectionCount)
		if hConnectionCount < hostConnectionCount {
			host = h
		}
	}
	return host
}
