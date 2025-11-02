# chaos-proxy

Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on. Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.

---

## Objective

Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on.

Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.

---

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

- [ ] Verify correct port-to-upstream mapping
- [ ] Test with multiple simultaneous connections
- [ ] Confirm each listener starts on correct port

**Stretch:**

- [ ] Per-route connection timeouts
- [ ] Rate limiting per route
- [ ] Connection pooling to upstreams

### 3. Data Forwarding

**Core:**

- [x] Establish new TCP connection to target upstream
- [x] Copy data bidirectionally until either side closes

**Testing/Verification:**

- [ ] Verify data passes through unchanged
- [ ] Test bidirectional data flow
- [ ] Confirm cleanup on connection close

**Stretch:**

- [ ] Track bytes transferred per connection
- [ ] Log data flow statistics
- [ ] Connection keep-alive support

### 4. Chaos Engineering

**Core:**

- [x] Implement `dropRate` (0.0–1.0 probability of dropping connection)
- [x] Implement `latencyMs` (artificial delay before forwarding begins)

**Testing/Verification:**

- [ ] Verify drop rate follows configured probability
- [ ] Confirm latency delay timing
- [ ] Test chaos behavior doesn't corrupt data

**Stretch:**

- [ ] Packet corruption/modification
- [ ] Bandwidth throttling
- [ ] Random connection resets
- [ ] Jitter (variable latency)

### 5. Bonus Features

**Core:**

- [ ] Log key events (connections, upstreams, bytes transferred, chaos events)
- [ ] Handle SIGINT/SIGTERM gracefully
  - [ ] Stop accepting new connections
  - [ ] Allow in-flight connections to complete

**Stretch:**

- [ ] Metrics/monitoring endpoint
  - [ ] Connection counts per route
  - [ ] Error rates and types
  - [ ] Uptime statistics
- [ ] Health checks
  - [ ] Periodic upstream health probes
  - [ ] Automatic upstream failover
  - [ ] Circuit breaker pattern

---

## Deliverables

- [ ] `main.go` implementing the described proxy behavior.
- [ ] Example configuration file demonstrating valid routes.
- [ ] `README.md` including:
  - [ ] Instructions for building and running locally.
  - [ ] Description of design choices and limitations.

---
