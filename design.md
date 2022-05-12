# TCP Load Balancer Design

## Terminology/Assumptions

- **Upstream**: A destination for TCP traffic. Could be an application running on HTTP, a gRPC server, or any other protocols using HTTP.
- **Host**: An individual TCP address a client would connect to to reach an upstream.

For example, upstream A is a REST server running on three machines. We would say upstream A has 3 hosts. A client would connect to the load balancer to try to reach upstream A, and the load balancer would appropriately route the client to one of the 3 hosts. (Assuming the client authenticated, passed authorization checks, etc.)

## Libraries

### Request forwarder

The request forwarding library will implement a least-connections load balancing algorithm to forward TCP connections to a collection of known hosts for a given upstream. A library client (like our server) would create a load balancer for a given upstream using the following signature:

```go
func NewLoadBalancer(hosts []*net.TCPAddr) *LoadBalancer
```

When created, the load balancer would create an internal map of each host to the number of connections the host is currently handling.

When the server gets a TCP connection, it will pass it to the request forwarding library to route it to the appropriate upstream:

```go
func (lb *LoadBalancer) HandleConnection(ctx context.Context, conn net.TCPConn) error
```

At a high level, the `HandleConnection` function will:

- Iterate over each of the hosts to find the one with the fewest connections
    - The counters used for connections will be atomic (e.g. protected from concurrent access with a mutex, or use `sync/atomic` functions) to avoid multithreaded race conditions
- Dial the host to initiate a TCP connection with it
    - Return an error if the connection can't be established
- Increment (again, thread safely) the connection count for the host
- In two separate goroutines, `io.Copy` the connection passed in to the connection to the upstream and vice versa
- Block for each goroutine to signal its completion
- After completion, close both connections
- Decrement the connection count for the host

### Questions/Assumptions

- How to handle context cancelation/the passed-in connection timing out?
- Should we set a timeout on the connection to upstream?
- Do we need to handle errors in the goroutines?

### Client Rate Limiter

This library will implement a simple, in-memory rate limiter for client requests. Users will create a rate limiter by passing in a map of client identifier -> # of requests allowed per second like so:

```go
func NewRateLimiter(requestsPerClient map[string]int) *RateLimiter
```

(To keep things simple, we'll only allow per-second limits rather than customizing the timeframe used)

Consumers of the library can find out whether a request is allowed using the `isRequestAllowed` function:

```go
func (rl *RateLimiter) IsRequestAllowed(ctx context.Context, clientId string) (bool, error)
```

Internally the library will use [the token bucket algorithm](http://intronetworks.cs.luc.edu/current/html/tokenbucket.html) to rate limit requests. At creation it will construct a bucket for each client. Calling `IsRequestAllowed` will trigger the library to query the corresponding bucket and determine whether the client has enough tokens (always 1 for simplicity's sake) to make the request.

The bucket will use a mutex to synchronize access to its token count check to prevent multithreading issues where quick requests from the same client don't accurately read/modify the value. The rate the bucket refills will be hard coded to keep things simple, though in reality this would be configurable.


## Server

