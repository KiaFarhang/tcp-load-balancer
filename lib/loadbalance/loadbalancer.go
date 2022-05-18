// Package loadbalance implements a least-connections load balancer.
package loadbalance

import (
	"context"
	"io"
	"net"
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
LoadBalancer is a least-connections load balancer. It keeps track of the number
of connections to a group of hosts, and routes a request to whichever host has the fewest
at the time the request is processed.

If two hosts have the same number of connections, the LoadBalancer will always select whichever
host had the lower index in the list of hosts originally passed to it.
*/
type LoadBalancer struct {
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

Duplicate addresses are ignored; if two addresses passed share the same IP, zone and port they
will be treated as a single host when performing load balancing.
*/
func NewLoadBalancer(addresses []*net.TCPAddr) (*LoadBalancer, error) {
	hosts := make([]*host, 0, len(addresses))
	for _, address := range addresses {
		host := &host{address: address, connectionCount: &atomic.Counter{}}
		hosts = append(hosts, host)
	}

	// I believe when both a context and a dialer have a timeout the shorter value
	// is respected; this protects us from clients passing in a no-timeout context
	// and our dial deadlocking when we can't connect to the upstream.
	return &LoadBalancer{hosts, &net.Dialer{Timeout: maxConnectionTimeout}}, nil
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
func (lb *LoadBalancer) HandleConnection(ctx context.Context, conn net.Conn) {
	host := lb.findHostWithLeastConnections()
	connectionToHost, err := lb.dialer.DialContext(ctx, "tcp", host.address.String())

	if err != nil {
		select {
		case <-ctx.Done():
			conn.Write([]byte(connectionToUpstreamTimedOutMessage))
		default:
			conn.Write([]byte(internalServerErrorMessage))
		}
		conn.Close()
		return
	}

	host.connectionCount.Increment()
	defer host.connectionCount.Decrement()

	done := make(chan struct{})

	go func() {
		defer conn.Close()
		io.Copy(conn, connectionToHost)
		done <- struct{}{}
	}()

	go func() {
		defer connectionToHost.Close()
		io.Copy(connectionToHost, conn)
		done <- struct{}{}
	}()

	<-done
	<-done
}

func (lb *LoadBalancer) findHostWithLeastConnections() *host {
	host := lb.hosts[0]

	for _, h := range lb.hosts {
		if h.connectionCount.Get() < host.connectionCount.Get() {
			host = h
		}
	}

	return host
}
