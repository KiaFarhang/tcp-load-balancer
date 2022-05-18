# TCP Load Balancer Design

## Terminology/Assumptions

- **Upstream**: A destination for TCP traffic. Could be an application running on HTTP, a gRPC server, or any other protocols using HTTP.
- **Host**: An individual TCP address a client would connect to to reach an upstream.

For example, upstream A is a REST server running on three machines. We would say upstream A has 3 hosts. A client would connect to the load balancer to try to reach upstream A, and the load balancer would appropriately route the client to one of the 3 hosts. (Assuming the client authenticated, passed authorization checks, etc.)

## Libraries

### Request forwarder

The request forwarding library will implement a least-connections load balancing algorithm to forward TCP connections to a collection of known hosts for a given upstream. A library client (like our server) would create a load balancer for a given upstream using the following signature:

```go
func NewLoadBalancer(hosts []*net.TCPAddr) *Balancer
```

When created, the load balancer would create an internal map of each host to the number of connections the host is currently handling.

When the server gets a TCP connection, it will pass it to the request forwarding library to route it to the appropriate upstream:

```go
func (lb *Balancer) HandleConnection(ctx context.Context, conn net.Conn)
```

At a high level, the `HandleConnection` function will:

- Iterate over each of the hosts to find the one with the fewest connections
    - The counters used for connections will be atomic (e.g. protected from concurrent access with a mutex, or use `sync/atomic` functions) to avoid race conditions
- Dial the host to initiate a TCP connection with it
    - If the connection can't be established, write an error message to the `conn` passed in. Close the connection and return without further processing.
    - To avoid deadlocks, we'll use a timeout here to bail if the host doesn't respond in t seconds.
- Increment (again, thread safely) the connection count for the host
- Defer decrementing the connection count for the host
- In two separate goroutines, `io.Copy` the connection passed in to the connection to the upstream and vice versa
- Block for each goroutine to signal its completion
- After completion, close both connections

### Client Rate Limiter

This library will implement a simple, in-memory rate limiter for client requests. Users will create a rate limiter by passing in a map of client identifier -> # of requests allowed per second like so:

```go
func NewRateLimiter(requestsPerClient map[string]int) *RateLimiter
```

(To keep things simple, we'll only allow per-second limits rather than customizing the timeframe used)

Consumers of the library can find out whether a request is allowed using the `IsRequestAllowed` function:

```go
func (rl *RateLimiter) IsRequestAllowed(ctx context.Context, clientId string) (bool, error)
```

Internally the library will use [the token bucket algorithm](https://en.wikipedia.org/wiki/Token_bucket) to rate limit requests. At creation it will construct a bucket for each client. Calling `IsRequestAllowed` will trigger the library to query the corresponding bucket and determine whether the client has enough tokens (always 1 for simplicity's sake) to make the request.

The bucket will use a mutex to synchronize access to its token count to prevent multithreading issues where quick requests from the same client don't accurately read/modify the value. The rate the bucket refills will be hard coded to keep things simple, though in reality this would be configurable.

## Server

The server is responsible for authentication and authorization before handing off the main work to the load balancing library. Authentication will be done with mTLS, with certificates stored in the repo for the sake of simplicity. Authorization will also be hard coded; the server will keep a map of upstream port to clients allowed to access the upstream fronted by that port. 

For auth purposes we'll define "clients" as the email address in the SAN of the client cert. So our in-memory auth map might look something like this:

```go
allowedUpstreams := map[int][]string{7777: []string{"admin@example.com", "user@example.com"}, 8888: []string{"admin@example.com"}}
```

Upon receiving a request at port 7777 (say for upstream A), the server will read the email(s) in the cert's SAN. If any of them are in the list of allowed emails for that port, the request will be considered authorized.

This will let us use the same root CA for all clients but give us finer-grained control over which clients can access which upstreams.

The server will accept TCP connections on a dedicated port for each upstream - e.g. port 7777 for application A, port 8888 for application B. Whenever it receives a connection, it will spin up a goroutine to handle it like so:

- Read the client's identity from the SAN in the cert (e.g. `admin`, `user1`, `pikachu` - again, hard coded to keep things simple)
- Check the client identity against the hard-coded authorization data
    - If the client doesn't have access to the upstream for the port it's connected on, send it an error message, close its connection and return without further work
- Ask the rate limiter whether the client is allowed to make a request right now
    - If the client is rate limited, send it an error message, close its connection and return without further work
- Forward the client's connection to the load balancing library to pass it to an appropriate host for the upstream (or send it an error and close it if there's a problem connecting to an upstream)

## User Experience / Testing

Clients would be able to communicate with this load balancer the way they would for any TLS-secured TCP traffic. To simplify the review process and add test coverage, I'll write end-to-end tests for the application in Go.

The following scenarios are ranked in what I think are the highest priority. Depending on time constraints and reviewer interest, the last on the list may not be necessary.

1. Connecting to the load balancer on a port for upstream A correctly routes you to a host for upstream A (the upstream will just be some dummy TCP listeners created in the test)
2. Connecting to the load balancer on a port corresponding to an upstream the client isn't authorized for results in an error
3. If upstream A has a host with 1 active connection, another request to upstream A will route to a second host with 0 active connections
4. Making many paralell requests from the same client results in a rate-limited rejection (timing will make this hard to reproduce perfectly; will probably just throw a bunch of goroutines at it and assert at least one of them got the rate limiting error message)
5. Same as #1, but with a "hello world" HTTP upstream instead of a raw TCP listener

## Other Considerations

- For this exercise, we'll handle errors by simply writing an error message to the client's TCP connection and closing it. In a production system we'd want something more robust - perhaps custom, documented error codes, for example.
- This load balancer isn't scalable/couldn't be run redundantly without breaking the rate limiting and least-connection forwarding functionality, because they both run off values in memory. If, for example, we scaled to 3 load balancing servers, each would have its own count of how many connections a host had.
    - We could solve this by moving those values into a distributed cache, so each server instance operated with the same values. At a glance, [Redis supports distributed locks](https://redis.io/docs/reference/patterns/distributed-locks/) to ensure multiple clients can't collide when reading/writing to the cache. Other distributed caches likely do as well.
- Because the load balancer will simply `io.Copy` the connection streams back and forth to each other, by default the connections will stay open until one side closes them (or the load-balancing server shuts down). This is nice for long-running processes like a remote debugger, but in reality we'd probably want to provide the option to enforce connection timeouts.
    - This could also be seen as a security/performance risk. If I have access to upstream A, I could just `echo -n "ddos" | nc server xxxx` in a new process each time to keep tons of connections with upstream A open and hog resources. However, this assumes upstream A isn't configured to close incoming connections automatically after it responds. And ideally the rate limiting would mitigate some of this risk.
- Everything hard coded in this implementation could be improved by either reading it from a configuration file or at runtime:
    - Authorization could be delegated to a tool like [Open Policy Agent](https://www.openpolicyagent.org/)
    - Available upstreams and their corresponding hosts could be dynamically changed at runtime, or even discovered with some kind of service discovery mechanism
    - Rate limiting rules could be managed by a separate application, stored in a database and periodically queried + cached by the rate limiting library
- The proposed function signatures include a `context.Context` because in a real-world application we'd likely use them for cancellations, deadlines, etc. I don't plan on using them much in this code, but figured it's safer to take the argument in case that changes.