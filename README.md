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

- [x] Implement `dropRate` (0.0–1.0 probability of dropping connection)
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
- [ ] Example configuration file demonstrating valid routes.
- [ ] `README.md` including:
  - [ ] Instructions for building and running locally.
  - [ ] Description of design choices and limitations.

---
