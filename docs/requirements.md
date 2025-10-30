# Take-Home Assignment: Layer-4 Chaos Proxy

## Objective

Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on.  
Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.

---

## Functional Requirements

### 1. Port-based routing

- The proxy reads a configuration file (JSON or YAML) that lists routes.  
  Example:

```json
  [
    {
      "localPort": 8080,
      "upstream": "10.0.0.5:80",
      "dropRate": 0.1,
      "latencyMs": 200
    },
    {
      "localPort": 9090,
      "upstream": "10.0.0.6:80"
    }
  ]
````

- The proxy listens on each `localPort`.
- Each incoming TCP connection is forwarded to the corresponding `upstream`.

### 2. Data forwarding

- For each connection:

  - Establish a new TCP connection to the target.
  - Copy data bidirectionally until either side closes.

### 3. Chaos behavior

- `dropRate` — probability (0.0–1.0) that a connection is dropped instead of proxied.
- `latencyMs` — artificial delay before forwarding begins.

### 4. Configuration validation

- Duplicate local ports must be rejected at startup.
- Invalid JSON/YAML should cause an immediate error.

### 5. Bonus points

- Log key events: accepted connection, chosen upstream, bytes transferred, injected delays.
- Handle SIGINT/SIGTERM and stop accepting new connections while allowing in-flight ones to finish.

---

## Deliverables

- `main.go` implementing the proxy.
- Example configuration file.
- `README.md` with:

  - How to build and run locally.
  - Description of design choices and limitations.

---
