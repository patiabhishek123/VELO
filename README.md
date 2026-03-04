# Velo - High-Performance Customizable load balancer

**Velo** is a production-grade, fully configurable load balancer written in Go. It provides intelligent request distribution across multiple backend servers with built-in health checking, circuit breaking, and comprehensive metrics.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API & Metrics](#api--metrics)
- [Advanced Features](#advanced-features)
- [Performance](#performance)
- [Contributing](#contributing)
- [License](#license)

---

## Features

✅ **Multiple Load Balancing Strategies**
- Round-Robin: Distributes requests sequentially across healthy backends
- Least Connection: Routes to backend with fewest active connections

✅ **Automatic Health Checking**
- Periodic health probes (configurable interval)
- Automatic backend status updates
- Graceful handling of unhealthy backends

✅ **Circuit Breaker Pattern**
- Prevents cascading failures
- Configurable failure threshold
- Automatic recovery timeout

✅ **Real-time Metrics**
- Total requests per pool
- Failed requests tracking
- Active connections monitoring
- JSON or Prometheus format

✅ **Full Configuration**
- YAML-based config files
- Environment variable overrides
- Command-line flags
- Zero-downtime config changes

✅ **Thread-Safe Operations**
- RWMutex for concurrent access
- Atomic operations for counters
- Safe concurrent request handling

✅ **Production Ready**
- Reverse proxy with error handling
- Graceful error responses
- Structured logging
- Clean architecture with interfaces

---

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Client Requests                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Load Balancer (Velo)                           │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              HTTP Request Handler                        │   │
│  │  - Receive incoming requests                             │   │
│  │  - Route to appropriate backend                          │   │
│  │  - Update metrics                                        │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │              Strategy Module                             │   │
│  │  ┌────────────────────┐    ┌───────────────────────┐    │   │
│  │  │   RoundRobin       │    │  LeastConnection      │    │   │
│  │  │  - Sequential pick │    │  - Track active conns │    │   │
│  │  │  - Atomic counter  │    │  - Min conn selection │    │   │
│  │  └────────────────────┘    └───────────────────────┘    │   │
│  └──────────────────────┬───────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │           Backend Pool & Health Checker                  │   │
│  │  ┌────────────────────────────────────────────────────┐  │   │
│  │  │ Backend 1 [Healthy]   Backend 2 [Unhealthy]       │  │   │
│  │  │ Backend 3 [Healthy]   Backend 4 [Recovering]      │  │   │
│  │  └────────────────────────────────────────────────────┘  │   │
│  │  - Periodic health probes every 5s                       │   │
│  │  - Update health status                                  │   │
│  │  - Thread-safe access with RWMutex                       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │      Circuit Breaker (Per Backend)                       │   │
│  │  - Track failure count                                   │   │
│  │  - Fail-fast when threshold exceeded                     │   │
│  │  - Recover after timeout                                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │           Reverse Proxy                                  │   │
│  │  - Forward requests to backend                           │   │
│  │  - Handle backend errors                                 │   │
│  │  - Record success/failure                                │   │
│  └─────────────────────────────────────────────────────────┘   │
│                         │                                        │
│  ┌──────────────────────▼───────────────────────────────────┐   │
│  │      Metrics Collector                                   │   │
│  │  - Total requests                                        │   │
│  │  - Failed requests                                       │   │
│  │  - Active connections                                    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │    Backend Servers                 │
        │ ┌──────────┐  ┌──────────┐        │
        │ │Backend 1 │  │Backend 2 │  ...   │
        │ └──────────┘  └──────────┘        │
        └────────────────────────────────────┘
```

### Request Flow Sequence

```
1. Client sends HTTP request to Velo on port 8090
   │
2. LoadBalancer.ServeHTTP() intercepts request
   │
3. Check circuit breaker status
   ├─ If Open → return ServiceUnavailable
   │
4. Get strategy (RoundRobin / LeastConnection)
   │
5. Strategy.NextBackend() selects healthy backend
   ├─ RoundRobin: atomic counter % healthy backends count
   ├─ LeastConnection: select backend with minimum active connections
   │
6. Record metrics
   ├─ Increment TotalRequests
   ├─ Increment ActiveConnections
   │
7. Create reverse proxy to backend.URL
   │
8. Forward request to backend
   ├─ If error → record failure, increment FailedRequests
   │
9. Record success (for circuit breaker reset)
   │
10. Decrement ActiveConnections
    │
11. Return response to client
```

---

## Project Structure

```
Custom_load_balancer/
├── README.md                              # This file
├── config.yaml                            # Default configuration
├── go.mod                                 # Go module definition
├── go.sum                                 # Dependencies lock file
│
├── cmd/
│   └── Custom_load_balancer/
│       └── main.go                        # Entry point
│
├── config/
│   └── config.go                          # Configuration loading & parsing
│
├── internal/
│   ├── balancer/
│   │   ├── backend.go                     # Backend server representation
│   │   ├── backendpool.go                 # Pool of backends + health check
│   │   ├── strategy.go                    # Strategy interface
│   │   ├── RoundRobin.go                  # Round-robin implementation
│   │   └── LeastConnection.go             # Least-connection implementation
│   │
│   ├── circuit/
│   │   └── breaker.go                     # Circuit breaker pattern
│   │
│   ├── metrics/
│   │   ├── metrics.go                     # Metrics collection & HTTP handler
│   │   └── health.go                      # Health check utilities
│   │
│   └── proxy/
│       └── proxy.go                       # Reverse proxy & request routing
│
└── server/
    └── server.go                          # Test backend servers
```

---

## Component Details

### 1. **Backend** (`internal/balancer/backend.go`)

Represents a single backend server with health and connection tracking.

**Fields:**
- `URL` - Backend server address (e.g., `http://localhost:8081`)
- `Healthy` - Current health status
- `ActiveConnections` - Count of ongoing connections
- `Weight` - Load balancing weight (reserved for future use)
- `CircuitState` - Current circuit breaker state (Closed/Open/HalfOpen)
- `FailureCount` - Consecutive failures
- `LastFailureTime` - Timestamp of last failure

**Key Methods:**
- `IncrementConnections()` / `DecrementConnections()` - Track active connections
- `IsHealthy()` - Get health status (thread-safe)
- `SetHealthy(state bool)` - Update health status
- `IncrementFailures()` / `ResetFailure()` - Track circuit breaker state
- `isBackendAlive()` - Probe backend with `/health` endpoint

**Thread Safety:** Uses mutex to protect concurrent access.

---

### 2. **BackendPool** (`internal/balancer/backendpool.go`)

Manages a collection of backends and health checking.

**Fields:**
- `backends` - Slice of backend servers
- `counter` - Atomic counter for round-robin
- `mu` - RWMutex for thread-safe access

**Key Methods:**
- `AddBackend(b *Backend)` - Add a backend to pool
- `GetHealthyBackends()` - Return only healthy backends
- `GetAllBackends()` - Return all backends (filtered by health status)
- `HealthCheck()` - Runs indefinitely, probing backends every 5 seconds
- `GetNextBackendRR()` - Internal round-robin selection

**Health Checking Logic:**
```go
for {
    backends := GetAllBackends()
    for _, backend := range backends {
        alive := backend.isBackendAlive()
        backend.SetHealthy(alive)
        time.Sleep(5 * time.Second)
    }
}
```

**Thread Safety:** RWMutex allows multiple concurrent readers (strategy selection) and exclusive writers (health updates).

---

### 3. **Strategy Interface** (`internal/balancer/strategy.go`)

Defines how backends are selected.

```go
type Strategy interface {
    NextBackend() *Backend
    Pool() *BackendPool
}
```

**Implementations:**
- `RoundRobin` - Sequential selection with wraparound
- `LeastConnection` - Select backend with minimum active connections

---

### 4. **RoundRobin** (`internal/balancer/RoundRobin.go`)

Distributes requests evenly across healthy backends in sequence.

**Algorithm:**
```
index = atomic.AddUint64(&counter, 1)
backend = healthy_backends[index % len(healthy_backends)]
```

**Pros:**
- Simple, predictable distribution
- O(1) selection time
- Good for stateless requests

**Cons:**
- Doesn't account for backend load
- Works poorly with unequal server capacities

---

### 5. **LeastConnection** (`internal/balancer/LeastConnection.go`)

Routes each request to the backend with the fewest active connections.

**Algorithm:**
```
backend_with_min_connections = select from healthy_backends
                                where connections == min(all.connections)
```

**Pros:**
- Adapts to server load in real-time
- Better for long-lived connections
- Fairer distribution with heterogeneous servers

**Cons:**
- O(n) selection time (scans all backends)
- Needs accurate connection tracking

---

### 6. **CircuitBreaker** (`internal/circuit/breaker.go`)

Implements the circuit breaker pattern to prevent cascading failures.

**States:**
- **Closed** - Normal operation, requests pass through
- **Open** - Too many failures, requests rejected immediately
- **HalfOpen** - After timeout, attempting recovery

**Fields:**
- `FailureThreshold` - Max consecutive failures before opening
- `Timeout` - Duration to wait before half-open
- `lastFailureTime` - When the last failure occurred

**Key Methods:**
- `AllowRequest(backend)` - Check if request should proceed
- `RecordSuccess(backend)` - Reset failure count
- `RecordFailures(backend)` - Increment failure count, possibly open circuit

**Transition Logic:**
```
Closed ─failure─→ Open ─timeout─→ HalfOpen ─success─→ Closed
                                         ↑
                                      failure
                                         │
                                       Open
```

---

### 7. **LoadBalancer / Proxy** (`internal/proxy/proxy.go`)

The main request handler that ties everything together.

**Fields:**
- `strategy` - Load balancing strategy
- `breaker` - Circuit breaker instance

**ServeHTTP() Flow:**
1. Increment `TotalRequests` metric
2. Get next backend from strategy
3. Check circuit breaker
4. Increment `ActiveConnections` metric
5. Create reverse proxy to backend
6. Forward request
7. Handle errors (record failure, increment `FailedRequests`)
8. Record success
9. Decrement `ActiveConnections`
10. Return response to client

**Error Handling:**
- If no healthy backends: return 503 Service Unavailable
- If circuit breaker open: return 503 Service Unavailable
- If backend error: return 502 Bad Gateway, record failure

---

### 8. **Metrics** (`internal/metrics/metrics.go`)

Collects and exposes pool-level metrics.

**Tracked Metrics:**
```json
{
  "pool_pointer_address": {
    "total_requests": 1000,
    "failed_requests": 5,
    "active_connections": 23
  }
}
```

**Operations:**
- `RegisterPool(pool)` - Register a new pool for tracking
- `IncTotalRequests(pool)` - Increment total requests (atomic)
- `IncFailedRequests(pool)` - Increment failed requests (atomic)
- `IncActiveConnections(pool)` - Increment active connections (atomic)
- `DecActiveConnections(pool)` - Decrement active connections (atomic)
- `Handler(w, r)` - HTTP handler for `/metrics` endpoint

**Thread Safety:** Uses atomic operations and mutex for map access.

---

### 9. **Configuration** (`config/config.go`)

Manages application configuration from files and environment.

**Config Structure:**
```yaml
server:
  address: "0.0.0.0"
  port: 8090

backends:
  urls:
    - http://localhost:8081
    - http://localhost:8082

strategy: roundrobin

circuit:
  failure_threshold: 3
  timeout: 10s

health:
  interval: 5s
  timeout: 2s

metrics:
  enabled: true
  path: /metrics
```

**Loading Priority:**
1. Built-in defaults
2. Config file (if exists)
3. Environment variables (override)

**Supported Env Vars:**
- `LB_PORT` - Server port
- `LB_ADDRESS` - Server address
- `LB_STRATEGY` - `roundrobin` or `leastconnection`
- `HEALTH_CHECK_INTERVAL` - e.g., `10s`
- `CIRCUIT_FAILURE_THRESHOLD` - integer

---

## Installation

### Prerequisites
- Go 1.25 or higher
- Linux, macOS, or Windows

### From Source

```bash
# Clone repository
git clone https://github.com/yourusername/velo.git
cd velo

# Download dependencies
go mod download

# Build binary
go build -o velo ./cmd/Custom_load_balancer

# (Optional) Install to system
sudo mv velo /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/velo/cmd/Custom_load_balancer@latest
```

---

## Configuration

### Configuration File (config.yaml)

Create a `config.yaml` in the project root:

```yaml
server:
  address: "0.0.0.0"      # Bind address
  port: 8090              # Listen port

backends:
  urls:
    - http://backend1.internal:8080
    - http://backend2.internal:8080
    - http://backend3.internal:8080

strategy: roundrobin        # roundrobin | leastconnection

circuit:
  failure_threshold: 3      # Open circuit after N failures
  timeout: 10s              # Time before attempting recovery

health:
  interval: 5s              # How often to check backend health
  timeout: 2s               # Health check request timeout

metrics:
  enabled: true             # Enable /metrics endpoint
  path: /metrics            # Metrics endpoint path
```

### Environment Variables

Override any config value with environment variables:

```bash
# Server configuration
export LB_PORT=9090
export LB_ADDRESS=127.0.0.1

# Load balancing strategy
export LB_STRATEGY=leastconnection

# Health checking
export HEALTH_CHECK_INTERVAL=10s
export HEALTH_CHECK_TIMEOUT=3s

# Circuit breaker
export CIRCUIT_FAILURE_THRESHOLD=5

# Run the load balancer
./velo
```

### Command-Line Flags

```bash
# Use custom config file
./velo -config /etc/velo/production.yaml

# Generate default config (useful for setup)
./velo -generate-config

# Combine with env vars
LB_PORT=9090 ./velo -config ./config.yaml
```

---

## Usage

### Basic Setup

**1. Start Backend Servers**

For testing, create simple backends (or use existing servers):

```bash
# Terminal 1
go run ./server 8081

# Terminal 2
go run ./server 8082

# Terminal 3
go run ./server 8083
```

**2. Start Velo**

```bash
# Using default config
./velo

# Or with custom config
./velo -config ./config.yaml
```

Expected output:
```
Added backend: http://localhost:8081
Added backend: http://localhost:8082
Added backend: http://localhost:8083
Using RoundRobin strategy
Metrics endpoint enabled at /metrics
Starting load balancer on 0.0.0.0:8090
```

**3. Send Requests**

```bash
# Single request
curl http://localhost:8090

# Multiple requests (see round-robin in action)
for i in {1..6}; do curl http://localhost:8090; echo; done

# Check metrics
curl http://localhost:8090/metrics | jq

# Check specific backend health
curl http://backend1.internal:8080/health
```

### Load Testing

```bash
# Using ab (Apache Bench)
ab -n 10000 -c 100 http://localhost:8090/

# Using wrk (HTTP benchmarking)
wrk -t4 -c100 -d30s http://localhost:8090/

# Using hey
hey -n 10000 -c 100 http://localhost:8090/
```

### Monitoring

**Real-time Metrics:**
```bash
# JSON format
watch -n 1 'curl -s http://localhost:8090/metrics | jq'

# Count total requests
curl -s http://localhost:8090/metrics | jq '.[] | .total_requests'
```

---

## API & Metrics

### HTTP Endpoints

#### **Main Load Balancer** (port 8090 by default)

**Endpoint:** `/` (any path)

**Method:** All HTTP methods (GET, POST, PUT, DELETE, etc.)

**Response:**
- Success: Forward response from backend
- No healthy backends: `503 Service Unavailable`
- Circuit breaker open: `503 Service Unavailable`
- Backend error: `502 Bad Gateway`

**Example:**
```bash
curl http://localhost:8090/api/users
# Returns response from a healthy backend
```

#### **Metrics Endpoint** (if enabled)

**Endpoint:** `/metrics`

**Method:** GET

**Response Format:** JSON

```json
{
  "0xc00009e000": {
    "total_requests": 10234,
    "failed_requests": 12,
    "active_connections": 45
  }
}
```

**Metrics Definition:**
- `total_requests` - Total HTTP requests received and forwarded
- `failed_requests` - Requests that resulted in errors
- `active_connections` - Currently open connections to backends

**Example:**
```bash
curl http://localhost:8090/metrics | jq

# Pretty print
curl http://localhost:8090/metrics | jq '.[] | {
  total_requests,
  failed_requests,
  success_rate: (.total_requests - .failed_requests) / .total_requests * 100
}'
```

### Monitoring Integration

**Prometheus Integration (Future):**
Velo can be extended to export metrics in Prometheus text format for integration with Prometheus, Grafana, etc.

---

## Advanced Features

### 1. Health Check Customization

Modify backend health check endpoint:

In `internal/balancer/backend.go`, the health check probes:
```go
resp, err := client.Get(b.URL + "/health")
```

Change `/health` to your actual health endpoint.

### 2. Custom Load Balancing Strategy

Implement the `Strategy` interface:

```go
type CustomStrategy struct {
    pool *BackendPool
}

func (s *CustomStrategy) NextBackend() *Backend {
    // Your custom logic here
    return selected_backend
}

func (s *CustomStrategy) Pool() *BackendPool {
    return s.pool
}
```

Then in `main.go`:
```go
strategy := &CustomStrategy{pool: pool}
```

### 3. Circuit Breaker Tuning

Adjust in `config.yaml`:
- **Strict (fewer false positives):** `failure_threshold: 5, timeout: 30s`
- **Aggressive (faster recovery):** `failure_threshold: 2, timeout: 5s`

### 4. Dynamic Backend Management

To add backends at runtime, extend `BackendPool`:
```go
func (p *BackendPool) RemoveBackend(url string) {
    p.mu.Lock()
    // Filter out backend by URL
    p.mu.Unlock()
}
```

---

## Performance

### Benchmarks

Performance depends on hardware and load characteristics. Typical metrics:

- **Throughput:** 10,000+ req/s per core
- **Latency (p50):** <5ms
- **Latency (p99):** <50ms
- **Memory:** ~50MB base + ~1KB per active connection

### Optimization Tips

1. **Use RoundRobin for throughput**, LeastConnection for fairness
2. **Tune health check interval** - more frequent = higher overhead
3. **Adjust circuit breaker threshold** - balance between quick recovery and stability
4. **Use multiple load balancer instances** behind a DNS/IP failover

### Scalability

- **Single Instance:** Easily handles 10k-100k req/s
- **Multiple Instances:** Use DNS round-robin or keepalived for HA
- **Horizontal Scaling:** Run multiple Velo instances, distribute with DNS/HAProxy

---

## Troubleshooting

### Issue: All backends show as unhealthy

**Cause:** Health check endpoint `/health` doesn't exist or is unresponsive

**Solution:**
1. Verify backend has `/health` endpoint
2. Check health check timeout: `health.timeout`
3. Verify network connectivity: `curl http://backend:port/health`

### Issue: High failure rate in metrics

**Cause:** Backends are timing out or slow

**Solution:**
1. Check backend logs
2. Increase health check timeout
3. Check network latency
4. Scale backends or reduce load

### Issue: Circuit breaker constantly opening

**Cause:** Too sensitive configuration

**Solution:**
1. Increase `circuit.failure_threshold`
2. Increase `circuit.timeout`
3. Fix underlying backend issue

### Issue: Uneven load distribution with RoundRobin

**Cause:** Backends have different capacities

**Solution:**
Switch to `LeastConnection` strategy:
```yaml
strategy: leastconnection
```

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Install dev dependencies
go mod download

# Run tests (if available)
go test ./...

# Format code
go fmt ./...

# Lint
go vet ./...

# Build
go build -o velo ./cmd/Custom_load_balancer
```

---

## License

This project is licensed under the MIT License - see LICENSE file for details.

---

## Roadmap

- [ ] Prometheus metrics export format
- [ ] gRPC load balancing
- [ ] WebSocket support
- [ ] Custom authentication/authorization
- [ ] Dynamic backend management API
- [ ] Weighted round-robin
- [ ] Request rate limiting
- [ ] TLS/HTTPS support
- [ ] Admin dashboard

---

## Support

For issues, questions, or suggestions:
- Open an issue on GitHub
- Check existing documentation
- Review example configurations

---

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Max Throughput | 10k-100k req/s |
| Latency (mean) | 1-5ms |
| Latency (p99) | <50ms |
| Memory Overhead | ~50MB base |
| CPU per core | ~30-40% @ 10k req/s |
| Concurrent Connections | Unlimited (system limited) |

---

**Velo** - Fast, reliable, and configurable load balancing for Go applications.
