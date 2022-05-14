// Package loadbalance implements a least-connections load balancer
package loadbalance

import (
	"net"

	"github.com/KiaFarhang/tcp-load-balancer/lib/atomic"
)

// host is an individual server that can take traffic from the load balancer
type host struct {
	address         *net.TCPAddr
	connectionCount *atomic.Counter
}

/**
LoadBalancer is a least-connections load balancer. It keeps track of the number
of connections to a group of hosts, and routes a request to whichever host has the fewest
at the time the request is processed.

If two hosts have the same number of connections, the LoadBalancer will always select whichever
host had the lower index in the list of hosts originally passed to it.
*/
type LoadBalancer struct {
	hosts []*host
}

/**
NewLoadBalancer constructs a new least-connections load balancer to route
requests to the slice of TCP addresses provided. Clients should construct a separate
LoadBalancer for each upstream application they wish to load balance.
*/
func NewLoadBalancer(addresses []*net.TCPAddr) *LoadBalancer {
	var hosts []*host
	for _, address := range addresses {
		host := &host{address: address, connectionCount: &atomic.Counter{}}
		hosts = append(hosts, host)
	}

	return &LoadBalancer{hosts}
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
