# Progress Log

## 2025-11-03 - Documentation Polish, Example Configs & Test Server Feature

### Context 11-03

After completing core functionality, shifted focus to developer experience and interview presentation. Created comprehensive example configurations, implemented a test server feature to eliminate manual test setup, and rewrote the README to tell the story of design decisions and trade-offs rather than just list features.

### Example Configuration Library

**Created comprehensive config examples** to demonstrate different use cases and validation behavior:

**Valid configurations** (`examples/configs/valid/`):

- `basic.json` - Simple pass-through proxy (getting started)
- `multiple_routes.json` - Three services with progressive chaos (0%, 5%, 15% drop rates)
- `realistic_chaos.json` - Real-world network conditions (3%, 8%, 20% drop rates)
- `extreme_conditions.json` - Stress testing scenarios (75%, 90%, 100% drop rates)
- `latency_tiers.json` - Six latency tiers testing timeout tolerance (50ms to 3000ms)
- `ipv6.json` - IPv6 upstream example

**Invalid configurations** (`examples/configs/invalid/`) for testing validation:

- `duplicate_ports.json` - Port uniqueness validation
- `invalid_upstream_hostname.json` - IP-only requirement
- `invalid_port_range.json` - Port range bounds checking
- `invalid_drop_rate.json` - Drop rate bounds (0.0-1.0)
- `invalid_negative_latency.json` - Non-negative latency requirement
- `invalid_missing_upstream.json` - Required field validation
- `invalid_zero_port.json` - Static port assignment requirement

**Port selection**: Avoided common ports like 8080 (often in use). Chose 8180-8186 range for proxy listeners to reduce port conflicts during testing.

### Test Server Feature Implementation

**Problem identified**: Testing required manually starting upstream servers with `nc` or Python's HTTP server before each test run. This created friction during development and demonstration.

**Solution**: Added `-test-server` flag that automatically spins up simple HTTP servers on all configured upstream addresses.

**Implementation details**:

- Created `internal/testserver/testserver.go` with `NewTestServer()` function
- Uses `http.NewServeMux()` to create isolated handlers (avoids global handler conflicts)
- Runs each test server in a goroutine (non-blocking, don't track with WaitGroup)
- Test servers log to `slog` like the rest of the application
- Simple response includes timestamp and "under heavy load (of self-doubt and coffee)" message

**Key learning**: Initially started test servers synchronously, which blocked proxy listener startup. Fixed by launching in goroutines with a brief sleep to let servers bind before proxy starts. This demonstrates the same concurrency pattern used in the main proxy logic.

### README Transformation (Interview-Focused)

**Rewrote design section** to be reviewer-focused rather than feature-focused. Changed from listing what was built to explaining *why* decisions were made and *what limitations* resulted.

**New structure**:

- Problem/Solution/Trade-off/Limitation format for each design area
- Explicit callouts of production gaps (no connection pooling, no backpressure, shutdown can block indefinitely)
- Testing limitations acknowledged (no load testing, no race detector in CI, statistical tests could flake)
- Real-world constraints documented (no hot reload, no metrics, config changes require restart)

**Rationale**: Interview documentation should demonstrate self-awareness and systems thinking. Better to show honest understanding of limitations than pretend they don't exist. This signals senior-level reasoning about trade-offs.

**Technical improvements**:

- Added "Getting Started" section with prerequisites, build, and run instructions
- Documented all CLI flags (including new `-test-server`)
- Provided testing examples using the test server feature
- Updated configuration section with detailed field descriptions
- Marked all functional requirements as complete

### Key Learnings & Reflections

**Developer experience matters**: The `-test-server` flag was added as a convenience during development but turned out valuable enough to document as a user-facing feature. Good tools often emerge from scratching your own itch.

**Documentation is code review**: Writing honest documentation about limitations forces you to think through production concerns you might otherwise miss. Documenting "why not" is as important as documenting "why."

**Iteration continues**: Even after "completing" requirements, there's value in improving ergonomics, documentation, and presentation. The difference between "works" and "polished" is often the difference between junior and senior execution.

---

## 2025-11-02 - Graceful Shutdown Implementation

### Context 11-02

Completed the final bonus requirement of implementing graceful shutdown handling for SIGINT/SIGTERM signals. This work addressed the open issue identified on 11-01 where the program used `wg.Wait()` which blocks indefinitely. The implementation uses Go's `context` package to propagate cancellation signals throughout the program architecture.

### Graceful Shutdown Implementation

**Signal handling**: Implemented graceful shutdown using `context.Context` to coordinate shutdown across all goroutines:

- Created root context in `cmd/main.go` that listens for SIGINT/SIGTERM signals
- Passed context down through the program hierarchy (main → route listeners → connection handlers)
- Each goroutine watches for context cancellation and exits cleanly when signaled
- All listeners and active connections close gracefully when shutdown signal is received

**Context propagation pattern**:

- Root context created with `signal.NotifyContext()` for automatic signal handling
- Context passed to `ListenAndServeRoute()` functions
- Context propagated to `handleConnection()` goroutines
- Each level checks `ctx.Done()` to determine when to exit

**Design decision**: Kept implementation simple and straightforward. Initially considered a grace period approach but avoided adding complexity with multiple done channels and additional goroutines that could create confusion and potential resource leaks.

> **Post-implementation review note**: After reviewing requirements, discovered that the initial implementation actually overdid the graceful shutdown. It was forcefully closing active connections immediately when the context was cancelled, rather than allowing them to complete naturally. The requirement explicitly calls for "allowing in-flight connections to complete" rather than terminating them. Fixed by removing the goroutine that watched `ctx.Done()` and closed connections, allowing connections to finish naturally when either side closes or data transfer completes. This is a good reminder to validate implementation against requirements, not just implement what seems like the "right" behavior.

### Log Level Refinements

**Updated log levels** to be more appropriate for production use:

- Refined which messages appear at each log level
- Ensured quiet/verbose modes provide appropriate detail levels
- Improved overall logging clarity and usefulness

### Key Learnings

**Context package insights**:

- Multiple approaches possible for context-based cancellation, but simplicity is key
- Need to carefully evaluate complexity when adding goroutines and channels
- Context propagation through function signatures provides clean cancellation flow

**Concurrency lessons**:

- Go's concurrency is powerful but can quickly become complex if not carefully managed
- Multiple done channels and cancellation paths can lead to confusion and bugs
- Important to keep concurrency patterns simple and understandable

**Channel management**:

- Must be cautious of leaving channels undrained in concurrent code
- Undrained channels can cause goroutine leaks and unexpected behavior
- Careful resource management is critical in concurrent programs

**Refactoring approach**:

- Continually evaluate implementation approaches as code evolves
- Sometimes the simpler solution is the better solution, even if more sophisticated patterns exist
- Regular refactoring and reassessment helps prevent complexity creep

---

## 2025-11-01 - Proxy Implementation, Configuration Validation & Logging Refinement

### Context 11-01

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

### Configuration Testing Updates

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

### Testing & Proxy Refinements

**Proxy module testing** (`internal/proxy/proxy_test.go`):

- **Data forwarding tests**: Bidirectional data transfer verification (simple, empty, multiline, large messages)
- **Connection handling**: Multiple simultaneous connections, connection cleanup
- **Chaos behavior tests**:
  - Drop rate testing (deterministic 0.0/1.0, statistical 0.5)
  - Latency testing (0ms, 50ms, 100ms, 200ms delays)
  - Combined chaos scenarios
- **Upstream failure handling**: Tests behavior when upstream server is unreachable

**Main loop fix**: Updated `cmd/main.go` to properly wait for all goroutines using `sync.WaitGroup`:

```go
var wg sync.WaitGroup
for _, route := range routeConfigs {
    wg.Add(1)
    go func(r config.RouteConfig) {
        defer wg.Done()
        err := proxy.ListenAndServeRoute(r)
        // ...
    }(route)
}
wg.Wait()
```

This allows all listeners to start concurrently and remain open. Each route now runs in its own goroutine, fixing the sequential blocking issue.

**Bidirectional communication**: Updated `handleConnection()` to only wait for one side to close (`<-done` once) rather than both, aligning with requirements that connections should remain open until either side closes.

### Key Learnings & Open Issues

**Architectural issue resolved**: Fixed sequential route blocking by wrapping each `ListenAndServeRoute()` call in a goroutine with proper `sync.WaitGroup` synchronization.

**Design issue identified**: `handleConnection()` returns `error`, but since it's called with `go`, that return value is discarded into the void. Two potential approaches considered:

1. Change signature to not return error (just log internally)
2. Use error channels if we need to track connection failures

*Note: Decision on this approach was deferred until implementing structured logging (see logging section below).*

**Open issue resolved on 11-02**: Need to implement graceful shutdown handling (SIGINT/SIGTERM) rather than just cancellation. Currently using `wg.Wait()` which blocks indefinitely. *This was completed on 11-02 using context-based cancellation.*

**Debugging insight**: Used VS Code debugger to trace goroutine execution. Observed defer mechanism (connection cleanup happens before goroutine exit) and understood concurrency timing (main loop prints "Blocking..." before goroutine finishes printing its messages, because they run in parallel).

### Structured Logging Implementation

**Work continued from late 11-01 into early 11-02**: After resolving the proxy implementation issues, work shifted to implementing structured logging. This was a continuous effort spanning the end of 11-01 into the beginning of 11-02 to improve observability and error handling.

#### Logger Package Implementation

**Created `internal/logger/logger.go`** - Centralized logging configuration:

- Uses `slog` for structured logging with level control
- Supports three modes via flags:
  - **Default**: `LevelInfo` - Shows important events and errors
  - **Verbose** (`-verbose`): `LevelDebug` - Shows all detailed debug information
  - **Quiet** (`-quiet`): `LevelError` - Only shows errors
- Handles conflicting flags (both `-verbose` and `-quiet` set) with warning message, quiet takes precedence
- Custom `ReplaceAttr` function removes time from log output (can be re-enabled for production logging)

**Key learnings about slog**:

- **Contextual logging**: Using `slog.With()` to add default key-value pairs to loggers (e.g., `slog.With("file", configPath)` adds file context to all log messages)
- **Attribute filtering**: Ability to filter out default attributes like time using `ReplaceAttr` callback
- **Logger chaining**: Can extend loggers throughout program execution (e.g., config logger adds file context, route logger adds port context)
- **Structured attributes**: All log messages use key-value pairs for programmatic parsing and filtering

#### Enhanced Error Messages

**Updated all error logging** across the codebase to include helpful context:

- **Structured attributes**: Every error includes relevant context (file paths, route indices, port numbers, etc.)
- **Hint properties**: Added actionable `hint` attributes to error messages providing guidance on how to fix issues
- **Validation errors**: Enhanced config validation errors with:
  - Valid ranges (e.g., `"valid_range": "1-65535"`)
  - Specific values received (e.g., `"port": 99999`)
  - Examples of correct format (e.g., `"hint": "upstream must be in format 'ip:port' (e.g., '127.0.0.1:9090')"`)

**Error return pattern**: Refactored to return single error from validation functions while detailed errors are logged via slog. This provides best of both worlds - simple error checking for callers, detailed context in logs.

**Goroutine error handling decision**: Resolved the design issue identified earlier regarding `handleConnection()` error returns. After implementing structured logging, decided to change `handleConnection()` signature to not return error - errors are now logged internally via structured logging. This simplifies goroutine error handling since return values from goroutines are discarded anyway, and we get better observability through structured logs.

#### Logging Integration

**Updated all packages** to use structured logging:

- **`cmd/main.go`**: Uses slog for startup, config loading, and lifecycle events
- **`internal/config/`**: Logger passed through validation functions, includes file context
- **`internal/proxy/`**: Route logger includes port context, connection logger includes client address
- **Log levels refined**:
  - Changed "listener started successfully" from Info to Debug (too verbose for normal operation)
  - Changed "handling new connection" from Info to Debug (only needed for debugging)

#### Testing Updates for Logging

**Updated all test suites** to work with new logging:

- **`internal/proxy/proxy_test.go`**: Added `TestMain` to set up silent logger (LevelError) to avoid cluttering test output
- **`internal/config/config_test.go`**: Created `testLogger()` helper, updated validation tests to pass logger instead of config path string
- **`internal/logger/logger_test.go`**: Comprehensive test coverage:
  - Default/verbose/quiet mode behavior
  - Conflict handling (both flags set)
  - Time removal verification
  - Log level verification table-driven tests

All tests passing with new logging infrastructure.

---

## 2025-10-30 - Configuration System Complete

### Context 10-30

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
