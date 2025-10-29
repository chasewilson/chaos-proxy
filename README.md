# chaos-proxy
Build a simple **TCP proxy** in Go that forwards connections based on the **local port** they arrive on.   Each port maps to a different upstream target and can optionally inject network faults such as latency, drops, or timeouts.

---

## Core Requirements - From Project File

### Port-based routing
- [ ] Local listener ports map to configured upstream `host:port`.
  - [ ] Verify each configured local port has a corresponding upstream target.
  - [ ] Ensure startup fails if a local port is duplicated in the config.

### Configuration file
- [ ] Routes defined via a local config file (JSON or YAML as stated in the file).
  - [ ] Validate config parses successfully on startup.
  - [ ] Startup emits a clear error and exits on malformed config.

### Chaos knobs (per-route)
- [ ] Latency injection (configured `latencyMs`).
  - [ ] Confirm configured latency is applied to the route.
- [ ] Connection drop probability (configured `dropRate`).
  - [ ] Confirm connections are dropped according to the configured probability.
- [ ] Per-connection timeout (configured `timeoutMs`).
  - [ ] Confirm timeouts are enforced per connection as configured.

### Connection handling and forwarding
- [ ] Accept incoming TCP connections on configured local ports.
  - [ ] Confirm connections are forwarded to the configured upstream.
  - [ ] Confirm chaos knobs are applied per connection as configured.

### Logging and observable events (as described)
- [ ] Emit logs for connection events and applied chaos actions.
  - [ ] Confirm logs indicate when a connection is dropped, delayed, or timed out.

### Shutdown behavior
- [ ] Support graceful shutdown on termination signals.
  - [ ] Confirm new connections are stopped during shutdown and in-flight connections are allowed to finish or respect configured timeouts.

### Errors and startup validation
- [ ] Fail fast on bind or configuration errors with clear messages.
  - [ ] Confirm process does not continue with partial/invalid route bindings.

### Tests (basic)
- [ ] Include basic tests or smoke checks demonstrating core behavior described in the file.
  - [ ] Confirm tests exercise latency, drop, and timeout behaviors at a basic level.

---

## Stretch Goals (listed in the project file as intended fit / usage)

### Local development and test usage
- [ ] Usable as a local dev/test tool to point clients at the proxy instead of real services.
  - [ ] Confirm local usage reproduces flaky-link scenarios described in the file.

### Containerization (run in Docker / CI)
- [ ] Support running the proxy inside a container image (project file notes container use).
  - [ ] Confirm container startup using a provided config file behaves equivalently to local runs.

### Kubernetes deployment (project file mentions Kubernetes-friendly fit)
- [ ] Runnable in Kubernetes (Deployment or DaemonSet) per the project file.
  - [ ] Confirm the proxy can be started in-cluster and serve traffic to clients that target it.

---
