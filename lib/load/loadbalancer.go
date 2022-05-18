// Package load implements a least-connections load balancer.
package load

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/KiaFarhang/tcp-load-balancer/internal/atomic"
)

const (
	internalServerErrorMessage          string        = "Internal server error"
	connectionToUpstreamTimedOutMessage string        = "Timed out connecting to upstream"
	maxConnectionTimeout                time.Duration = 3 * time.Second
)

// host is an individual server that can take traffic from the load balancer.
type host struct {
	address         *net.TCPAddr
	connectionCount *atomic.Counter
}

/*
Balancer is a least-connections load balancer. It keeps track of the number
of connections to a group of hosts, and routes a request to whichever host has the fewest
at the time the request is processed.

If two hosts have the same number of connections, the Balancer will always select whichever
host had the lower index in the list of hosts originally passed to it.
*/
type Balancer struct {
	hosts  []*host
	dialer *net.Dialer
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
		return &Balancer{}, err
	}

	hosts := make([]*host, 0, len(validateAddresses))
	for _, address := range validateAddresses {
		host := &host{address: address, connectionCount: &atomic.Counter{}}
		hosts = append(hosts, host)
	}

	// I believe when both a context and a dialer have a timeout the shorter value
	// is respected; this protects us from clients passing in a no-timeout context
	// and our dial deadlocking when we can't connect to the upstream.
	return &Balancer{hosts, &net.Dialer{Timeout: maxConnectionTimeout}}, nil
}

/*
HandleConnection takes a TCP connection, finds a suitable host to handle it,
then connects to the host and streams data between the connection and host until
both sides of the connection are closed.

If the connection to the host fails, HandleConnection will write an error message
to the incoming net.Conn and close it. It will also close the net.Conn if the communication
with the host succeeds; callers of HandleConnection do not need to close the net.Conn
they pass in.
*/
func (lb *Balancer) HandleConnection(ctx context.Context, conn net.Conn) {
	host := lb.findHostWithLeastConnections()

	host.connectionCount.Increment()
	defer host.connectionCount.Decrement()

	connectionToHost, err := lb.dialer.DialContext(ctx, "tcp", host.address.String())

	if err != nil {
		select {
		case <-ctx.Done():
			conn.Write([]byte(connectionToUpstreamTimedOutMessage))
		default:
			conn.Write([]byte(internalServerErrorMessage))
		}
		closeErr := conn.Close()
		log.Printf("Error closing connection after dial failure. Dial error: %s, connection close error: %s", err.Error(), closeErr.Error())
		return
	}

	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	go func() {
		defer conn.Close()
		io.Copy(conn, connectionToHost)
		waitGroup.Done()
	}()

	go func() {
		defer connectionToHost.Close()
		io.Copy(connectionToHost, conn)
		waitGroup.Done()
	}()

	waitGroup.Wait()
}

func (lb *Balancer) findHostWithLeastConnections() *host {
	host := lb.hosts[0]

	for _, h := range lb.hosts {
		if h.connectionCount.Get() < host.connectionCount.Get() {
			host = h
		}
	}

	return host
}
