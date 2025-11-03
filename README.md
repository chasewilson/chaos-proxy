# chaos-proxy

## Overview

**chaos-proxy** is a TCP proxy written in Go that forwards connections based on the local port they arrive on. Each port maps to a different upstream target and can optionally inject network faults for chaos engineering testing.

This tool is designed for testing distributed systems under adverse network conditions. You can simulate connection drops, add artificial latency, and observe how your applications handle these scenarios.

---

## Getting Started

### Prerequisites

This project requires **Go 1.25 or later**. Check your version with:

```bash
go version
```

### Building

From the repository root, build the binary:

```bash
go build -o chaos-proxy ./cmd
```

### Running

Run the proxy with a configuration file:

```bash
./chaos-proxy -config examples/configs/valid/basic.json
```

Or try a more interesting scenario with multiple routes and chaos:

```bash
./chaos-proxy -config examples/configs/valid/multiple_routes.json -verbose
```

**Available flags:**

- `-verbose` - Enable debug-level logging for detailed output
- `-quiet` - Show errors only (suppresses informational messages)
- `-test-server` - Automatically start HTTP test servers on all upstream targets (useful for testing)

**Important notes:**

- Upstream targets must use IP addresses with ports (e.g., `127.0.0.1:9090` or `[::1]:9090` for IPv6). Hostnames like `localhost:9090` are rejected during configuration validation.
- Graceful shutdown is supported. When you send SIGINT (Ctrl+C) or SIGTERM, the proxy stops accepting new connections and allows active connections to complete naturally before exiting.

### Testing the Proxy

The easiest way to test is using the `-test-server` flag, which automatically starts HTTP test servers on all upstream targets defined in your configuration:

```bash
./chaos-proxy -config examples/configs/valid/multiple_routes.json -test-server -verbose
```

The `-test-server` flag will:

- Read your configuration file
- Automatically start an HTTP server on each upstream address
- Start the proxy listeners on each local port
- Allow you to test immediately without manually setting up servers

Then in another terminal, test the routes:

```bash
curl http://127.0.0.1:8180/  # Stable service (0% drop, no latency)
curl http://127.0.0.1:8181/  # Light chaos (5% drop, 100ms latency)
curl http://127.0.0.1:8182/  # Moderate chaos (15% drop, 500ms latency)
```

#### Manual Testing (without `-test-server`)

If you want to test with your own upstream servers:

**Step 1:** Start an upstream server (using netcat or Python):

```bash
# Option A: Netcat echo server
nc -l 127.0.0.1 9090

# Option B: Python HTTP server
python3 -m http.server 9090 --bind 127.0.0.1
```

**Step 2:** Run the proxy pointing to your upstream:

```bash
./chaos-proxy -config examples/configs/valid/basic.json -verbose
```

**Step 3:** Connect through the proxy:

```bash
# For netcat
nc 127.0.0.1 8180

# For HTTP server
curl http://127.0.0.1:8180/
```

#### Testing Chaos Parameters

Try different configurations to see chaos in action:

```bash
# Extreme conditions (75% drop rate, 5s latency)
./chaos-proxy -config examples/configs/valid/extreme_conditions.json -test-server -verbose
curl http://127.0.0.1:8180/  # Most requests will fail or be very slow

# Latency tiers (no drops, just delays)
./chaos-proxy -config examples/configs/valid/latency_tiers.json -test-server
curl http://127.0.0.1:8180/  # 50ms delay
curl http://127.0.0.1:8185/  # 3000ms delay
```

#### Running the Test Suite

To run the Go test suite:

```bash
go test ./...
```

Or with verbose output:

```bash
go test -v ./...
```

## Configuration

### File Format

Configurations are defined as JSON arrays of route objects. Each route specifies a local port to listen on, an upstream target, and optional chaos parameters:

```json
[
  {
    "localPort": 8180,
    "upstream": "127.0.0.1:9090",
    "dropRate": 0.0,
    "latencyMs": 0
  }
]
```

**Fields:**

- `localPort` (integer) - Port to listen on (1-65535)
- `upstream` (string) - Target server in `ip:port` format (IP addresses only)
- `dropRate` (float) - Probability of dropping connections (0.0 to 1.0)
- `latencyMs` (integer) - Artificial delay in milliseconds before forwarding data (0 or higher)

### Example Configurations

Comprehensive sample configuration files are provided in the `examples/configs/` directory. See `examples/configs/README.md` for detailed descriptions of each scenario.

**Valid configurations:**

- `valid/basic.json` - Simple pass-through proxy (getting started)
- `valid/multiple_routes.json` - Three services with progressive chaos levels
- `valid/realistic_chaos.json` - Real-world network conditions
- `valid/extreme_conditions.json` - Stress testing and worst-case scenarios
- `valid/latency_tiers.json` - Six latency tiers from 50ms to 3000ms
- `valid/ipv6.json` - IPv6 upstream example

**Invalid configurations** (useful for testing validation behavior):

- `invalid/duplicate_ports.json` - Duplicate local ports
- `invalid/invalid_upstream_hostname.json` - Hostname instead of IP address
- `invalid/invalid_port_range.json` - Port number outside valid range
- `invalid/invalid_drop_rate.json` - Drop rate outside 0.0-1.0 range
- `invalid/invalid_negative_latency.json` - Negative latency value
- `invalid/invalid_missing_upstream.json` - Empty upstream field
- `invalid/invalid_zero_port.json` - Port 0 not allowed

## Design Choices & Development Process

This section is written for reviewers. It explains what I built, why I built it that way, and how I adjusted course when new information surfaced. A day-by-day record lives in `docs/progress-log.md`.

### Configuration validation (IP-only, fail fast)

- **Problem**: During testing, `localhost:9090` worked fine with `net.Dial()`, but hostnames introduce DNS resolution as a variable. In chaos testing, connection failures need clear attribution—is it from injected chaos or DNS lookup/caching? Hostnames also behave differently across environments (local vs container vs cloud).
- **Solution**: Enforce IP-only upstreams using `net.ParseIP()` and validate `ip:port` format with `net.SplitHostPort()`. Supports both IPv4 and IPv6.
- **Trade-off**: Less convenient than hostnames, but eliminates DNS as a confounding variable in chaos experiments. Validation errors include examples and hints.
- **Limitation**: Can't use service discovery by name. Users must resolve IPs before writing config, which doesn't match real-world service mesh patterns where names are the contract.

### Concurrency model (simple, observable)

- **Initial bug**: Listeners started sequentially and blocked (only first route would start).
- **Solution**: Run each `ListenAndServeRoute()` in its own goroutine with a `sync.WaitGroup` to coordinate shutdown.
- **Error handling decision**: Rather than add error channels for every goroutine (complexity), pivoted to structured logging. Errors are logged with full context instead of returned. Simpler call sites, better observability.
- **Limitations**:
  - If any listener fails to bind (e.g., port in use), program exits via `os.Exit(1)`. No partial operation.
  - No connection pooling—each connection creates a new upstream connection. Under load, could exhaust file descriptors or upstream limits.
  - No backpressure if upstreams are slow or unresponsive.

### Graceful shutdown (requirement-driven correction)

- **Requirement**: Stop accepting new connections on SIGINT/SIGTERM but allow in-flight connections to complete.
- **Implementation**: Context-based cancellation closes listeners immediately but lets active `io.Copy` loops finish naturally.
- **Correction**: Initially force-closed connections on context cancel (too aggressive). Re-read requirements and progress notes, removed force-close. See commit c15ed382 and 2025-11-02 log entry.
- **Limitation**: No grace period timer. If an upstream hangs or a client never closes, shutdown blocks indefinitely. Could add a timeout, but chose simplicity over handling edge cases with additional goroutines and forced cleanup.

### Logging (structured, practical)

- **Approach**: Go's `slog` for structured, leveled logs with `-verbose` and `-quiet` flags.
- **Implementation**: Logger chaining via `slog.With()` adds context (file paths, ports, client addresses) at each layer. Validation errors include actionable `hint` fields.
- **Limitations**:
  - Removed timestamps in the `ReplaceAttr` function (helpful for testing, problematic for production log correlation).
  - Logs to stdout only—no rotation, remote aggregation, or sampling.
  - No performance testing under high connection volume; uncertain if structured logging becomes a bottleneck.
  - Trade-off: Stdlib simplicity over third-party integrations (OpenTelemetry, structured log shippers, metrics correlation).

### Testing strategy (deterministic + statistical)

- **Validation tests**: Table-driven tests covering happy paths and edge cases (port ranges, formats, duplicates, IPv6).
- **Proxy behavior tests**: Deterministic chaos (0.0 and 1.0 drop rates), statistical checks (0.5 drop rate over iterations), latency timing, bidirectional copy correctness.
- **Dev ergonomics**: `-test-server` flag auto-spins HTTP upstreams for fast iteration without manual setup.
- **Limitations**:
  - No integration tests against real services or containerized environments.
  - Statistical tests could flake under system load (probability-based assertions).
  - No load/stress testing, no race detector in CI (`-race` flag), no fuzz testing.
  - Trade-off: Focused on correctness and coverage over performance benchmarks and chaos-at-scale scenarios.

### Scope and trade-offs

- **In scope**: TCP proxying, connection drops, latency injection, structured logs, graceful shutdown, strict config validation.
- **Deferred**: Packet corruption/reordering, bandwidth throttling, jitter patterns, dynamic reconfiguration, metrics/observability endpoints, health checks, circuit breaking, retry logic.
- **Real-world limitations**:
  - Can't simulate nuanced network conditions (gradual degradation, bursty packet loss, asymmetric latency).
  - No runtime visibility into active connections or chaos events beyond log parsing.
  - Config changes require full restart (operational friction in production).
  - No ability to schedule chaos experiments, ramp failure rates gradually, or target specific connection patterns.
- **Rationale**: Ship a reliable, testable core with clear documentation rather than spread effort across half-implemented features. Demonstrates depth in fundamentals (concurrency, validation, testing) over breadth without quality.

## Future Evolution

### Kubernetes Network Chaos Testing

This project could evolve into a comprehensive network chaos testing tool for Kubernetes environments. The core TCP proxy functionality provides a solid foundation for more advanced chaos engineering scenarios in distributed systems.

**Potential enhancements:**

**Expanded chaos functionality:**

- Packet corruption and manipulation
- Bandwidth throttling and traffic shaping
- Network jitter and variable latency
- Partial connection failures (e.g., one-way communication loss)

**Kubernetes integration:**

- Deploy as sidecar containers alongside application pods
- Intercept service-to-service communication transparently
- Dynamic configuration via ConfigMaps or CRDs

**Observability and metrics:**

- Prometheus metrics export for connection counts, error rates, and chaos events
- Integration with observability platforms
- Real-time dashboards showing service impact under chaos conditions
- Correlation between injected faults and application behavior
- Automated chaos experiments with success criteria validation

**Use cases:**

- Test microservice resilience to network partitions and degraded connectivity
- Validate retry logic, circuit breakers, and timeout configurations
- Measure blast radius and cascading failure scenarios
- Continuous chaos testing in staging environments
- Pre-production validation before major deployments

This evolution would transform chaos-proxy from a standalone testing tool into a Kubernetes-native chaos engineering platform, enabling teams to build confidence in their distributed systems' resilience under realistic failure conditions.

## Functional Requirements

### 1. Configuration

**Core:**

- [x] Read configuration file (JSON or YAML) listing routes
- [x] Each route includes `localPort` and `upstream`
- [x] Reject duplicate `localPort` values at startup
- [x] Invalid JSON/YAML causes immediate error with clear message

**Testing/Verification:**

- [x] Confirm routes parse with expected fields
- [x] Verify startup behavior with invalid configurations
- [x] Test duplicate port detection

### 2. Port-based Routing

**Core:**

- [x] Listen on each `localPort` defined in configuration
- [x] Forward each incoming TCP connection to corresponding `upstream`

**Testing/Verification:**

- [x] Verify correct port-to-upstream mapping
- [x] Test with multiple simultaneous connections
- [x] Confirm each listener starts on correct port

### 3. Data Forwarding

**Core:**

- [x] Establish new TCP connection to target upstream
- [x] Copy data bidirectionally until either side closes

**Testing/Verification:**

- [x] Verify data passes through unchanged
- [x] Test bidirectional data flow
- [x] Confirm cleanup on connection close

### 4. Chaos Engineering

**Core:**

- [x] Implement `dropRate` (0.0 to 1.0 probability of dropping connection)
- [x] Implement `latencyMs` (artificial delay before forwarding begins)

**Testing/Verification:**

- [x] Verify drop rate follows configured probability
- [x] Confirm latency delay timing
- [x] Test chaos behavior doesn't corrupt data

### 5. Bonus Features

**Core:**

- [x] Log key events
  - [x] Connections
  - [x] Upstreams
  - [x] Bytes transferred
  - [x] Chaos events
- [x] Handle SIGINT/SIGTERM gracefully
  - [x] Stop accepting new connections
  - [x] Allow in-flight connections to complete

---

## Deliverables

- [x] `main.go` implementing the described proxy behavior.
- [x] Example configuration file demonstrating valid routes.
- [x] `README.md` including:
  - [x] Instructions for building and running locally.
  - [x] Description of design choices, trade-offs, and development process (interview-focused).

---

## test update to validate build

