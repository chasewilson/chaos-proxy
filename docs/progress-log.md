# Progress Log

## 2025-11-01 - Proxy Implementation & Configuration Validation Refinement

### Context

Started implementing the proxy functionality in `internal/proxy/proxy.go`. While testing with `localhost:9090` as upstream addresses, discovered that Go's `net.Dial()` expects an explicit address format. This led to implementing strict upstream validation to catch configuration errors at load time rather than at runtime.

### Proxy Implementation

- **Initial TCP listener implementation**: Created `ListenAndServeRoute()` function
  - Sets up TCP listener on configured local port
  - Implements Accept() loop to handle incoming connections
  - Spawns goroutine for each connection with `handleConnection()`
  
- **Connection handling**: Implemented `handleConnection()` function
  - Establishes TCP connection to upstream server
  - **Discovery**: `net.Dial()` requires proper `host:port` format, leading to validation work below
  
- **Debugging visibility**: Added comprehensive println statements throughout execution flow
  - Startup: Config loading, route enumeration
  - Listener lifecycle: Starting, blocking on Accept(), connection acceptance
  - Connection handling: Upstream dialing, success/failure states
  - This output critical for understanding sequential blocking issue (see Key Learnings)

### Configuration Validation Enhancement

**Problem discovered**: Test config used `localhost:9090` format, but realized this could mask issues. `net.Dial()` accepts hostnames, but for a chaos proxy, explicit IP addresses provide better control and clarity.

**Solution implemented**: Strict upstream validation using Go's `net` package

- Used `net.SplitHostPort()` to parse `host:port` format and validate structure
- Used `net.ParseIP()` to enforce IP addresses (reject hostnames/domains)
- Validates upstream port range (1-65535)
- Supports both IPv4 (`127.0.0.1:9090`) and IPv6 (`[::1]:9090`) formats
- Rejects common mistakes:
  - Hostnames: `localhost:9090`
  - URLs with schemes: `http://127.0.0.1:9090`
  - Missing ports: `127.0.0.1`
  - Invalid ports: `127.0.0.1:99999`

### Testing Updates

Updated all 46 existing test cases to use valid `ip:port` format instead of hostname-based addresses. Added 5 new test cases specifically for upstream validation:

- Hostname rejection test
- URL scheme rejection test  
- Missing port detection test
- Invalid port range test
- IPv6 address support verification

All tests passing.

### Development Tooling

- Created `.vscode/launch.json` for Go debugging sessions with proper `-config` argument
- Created `test-config.json` with valid IP-based configuration (3 routes: ports 8081-8083)

### Key Learnings & Open Issues

**Architectural issue identified**: Current implementation starts routes sequentially in a loop. `ListenAndServeRoute()` blocks on `Accept()`, so only the first route ever starts. Routes 2 and 3 never get initialized.

**Next step**: Need to wrap each `ListenAndServeRoute()` call in a goroutine to allow concurrent listeners.

**Design issue identified**: `handleConnection()` returns `error`, but since it's called with `go`, that return value is discarded into the void. Should either:

1. Change signature to not return error (just log internally)
2. Use error channels if we need to track connection failures

**Debugging insight**: Used VS Code debugger to trace goroutine execution. Observed defer mechanism (connection cleanup happens before goroutine exit) and understood concurrency timing (main loop prints "Blocking..." before goroutine finishes printing its messages, because they run in parallel).

---

## 2025-10-30 - Configuration System Complete

### Context

Built the configuration loading and validation system. This work focused on catching invalid configurations early and providing clear error messages to users.

### Core Implementation

**Configuration loading** (`LoadConfig` function):

- Reads JSON file and decodes into `[]RouteConfig`
- Uses `decoder.DisallowUnknownFields()` to catch typos/invalid fields
- Aggregates all validation errors (doesn't fail on first error)
- Returns both parsed config and error slice for caller handling

**Validation architecture** - Two-level validation approach:

1. `validateRouteConfig()`: Validates individual route fields
   - Port ranges (1-65535, zero explicitly rejected for static port requirement)
   - Drop rate bounds (0.0-1.0)
   - Latency non-negative
   - Upstream not empty (format validation added later on 11-01)

2. `validateConfig()`: Cross-route validation
   - Duplicate port detection using map lookup (O(n) instead of nested loop approach)
   - Accumulates per-route errors plus duplicate port errors

### Validation Logic Refinement

**Initial approach**: Considered nested loop to find duplicate ports, but realized this would be O(n²) and create redundant error messages.

**Better approach**: Use a map to track seen ports in single pass. This gives O(n) performance and reports each duplicate exactly once with its route index.

```go
portMap := make(map[int]struct{})
for i, route := range routes {
    if _, exists := portMap[route.LocalPort]; exists {
        // Report duplicate with route index
    }
    portMap[route.LocalPort] = struct{}{}
}
```

### Testing

Comprehensive test coverage across multiple test functions:

- `TestLoadConfig`: File I/O, JSON parsing, basic validation flow
- `TestLoadConfig_ValidationErrors`: All field validation rules with edge cases (boundary testing for ports 1 and 65535, drop rates 0.0 and 1.0)
- `TestLoadConfig_DuplicatePorts`: Duplicate detection with 2 and 3 route scenarios
- `TestValidateConfig`: Multi-route validation logic
- `TestValidateRouteConfig`: Individual field validation in isolation

Total: 46 test cases covering happy paths, sad paths, and edge cases.

### Documentation

- Updated README.md with completed configuration checklist items
- Validation rules clearly documented in code comments

---

## 2025-10-29 - Project Setup & Configuration Foundation

### Initial Setup

- Initialized Go module: `github.com/chasewilson/chaos-proxy`
- Established project structure following Go conventions:
  - `cmd/main.go` - Application entry point with CLI argument parsing
  - `internal/config/` - Configuration package (internal prevents external imports)
  - `docs/` - Project documentation and requirements
- Added `.gitignore` for Go build artifacts and IDE files

### Configuration Design

Defined `RouteConfig` struct to represent a single proxy route:

- `LocalPort int` - Port to listen on for incoming connections
- `Upstream string` - Target server address to forward connections to
- `DropRate float64` - Chaos parameter: probability (0.0-1.0) of dropping connections
- `LatencyMs int` - Chaos parameter: artificial delay in milliseconds

Design supports multiple routes via JSON array, allowing one proxy instance to handle multiple forwarding rules.

### Requirements Documentation

Created `docs/requirements.md` with:

- Feature checklist for tracking implementation progress
- MVP scope definition (basic proxy + chaos parameters)
- Deferred features (advanced chaos scenarios, metrics, etc.)

### Command-Line Interface

Set up basic CLI using `flag` package:

- `-config` flag for specifying configuration file path
- Validation that config path is provided (fail fast with clear error)
