# chaos-proxy
Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on.   Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.

---

# go TCP Chaos Proxy — Requirements Checklist

## Objective
Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on.  
Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.

---

## Functional Requirements

### 1. Port-based routing
- [ ] The proxy reads a configuration file (JSON or YAML) that lists routes.
  - [ ] Confirm routes include `localPort` and `upstream` fields.
  - [ ] Validate configuration parses correctly at startup.
- [ ] The proxy listens on each `localPort` defined in the configuration.
  - [ ] Confirm each listener starts successfully.
- [ ] Each incoming TCP connection is forwarded to the corresponding `upstream`.
  - [ ] Verify correct routing behavior per port mapping.

### 2. Data forwarding
- [ ] For each connection:
  - [ ] Establish a new TCP connection to the target `upstream`.
  - [ ] Copy data bidirectionally between client and upstream until either side closes.
  - [ ] Confirm data integrity is maintained during forwarding.

### 3. Chaos behavior
- [ ] Implement `dropRate` — probability (0.0–1.0) that a connection is dropped instead of proxied.
  - [ ] Verify that drop behavior follows configured probability.
- [ ] Implement `latencyMs` — artificial delay before forwarding begins.
  - [ ] Verify that latency delay occurs prior to upstream connection establishment.

### 4. Configuration validation
- [ ] Reject duplicate `localPort` values at startup.
  - [ ] Confirm startup fails if duplicate ports exist.
- [ ] Detect invalid JSON or YAML configuration.
  - [ ] Confirm invalid config causes immediate startup error with a clear message.

### 5. Bonus points
- [ ] Log key events:
  - [ ] Accepted connections.
  - [ ] Chosen upstream targets.
  - [ ] Bytes transferred.
  - [ ] Injected delays or dropped connections.
- [ ] Handle SIGINT/SIGTERM signals.
  - [ ] Stop accepting new connections upon signal.
  - [ ] Allow in-flight connections to complete before exiting.

---

## Deliverables
- [ ] `main.go` implementing the described proxy behavior.
- [ ] Example configuration file demonstrating valid routes.
- [ ] `README.md` including:
  - [ ] Instructions for building and running locally.
  - [ ] Description of design choices and limitations.

---