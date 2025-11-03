# Example Configurations

This directory contains example configuration files demonstrating both valid and invalid proxy configurations.

## Valid Configurations

### `valid/basic.json`

**Use case:** Simple pass-through proxy with no chaos  
**Routes:** 1 route on port 8180  
**Chaos:** None (0% drop, 0ms latency)  
**Best for:** Getting started, verifying basic proxy functionality

### `valid/multiple_routes.json`

**Use case:** Multiple services with varying chaos levels  
**Routes:** 3 routes (8180-8182)  
**Chaos:** Progressive chaos from none to moderate

- Port 8180: Stable (0% drop, 0ms latency)
- Port 8181: Light chaos (5% drop, 100ms latency)
- Port 8182: Moderate chaos (15% drop, 500ms latency)

**Best for:** Testing multiple services simultaneously with different reliability characteristics

### `valid/realistic_chaos.json`

**Use case:** Real-world network conditions  
**Routes:** 3 routes (8180-8182)  
**Chaos:** Realistic network degradation scenarios

- Port 8180: Good network (3% drop, 25ms latency)
- Port 8181: Degraded network (8% drop, 150ms latency)
- Port 8182: Poor network (20% drop, 500ms latency)

**Best for:** Testing under conditions similar to real production environments

### `valid/extreme_conditions.json`

**Use case:** Stress testing and worst-case scenarios  
**Routes:** 3 routes (8180-8182)  
**Chaos:** Extreme failure conditions

- Port 8180: Severe packet loss (75% drop, 5s latency)
- Port 8181: Nearly unusable (90% drop, 10s latency)
- Port 8182: Complete failure (100% drop)

**Best for:** Validating error handling, timeouts, and circuit breakers under extreme conditions

### `valid/latency_tiers.json`

**Use case:** Testing latency tolerance without connection drops  
**Routes:** 6 routes (8180-8185)  
**Chaos:** Pure latency testing (no drops)

- Port 8180: 50ms (fast local network)
- Port 8181: 100ms (typical LAN)
- Port 8182: 250ms (cross-datacenter)
- Port 8183: 500ms (high latency)
- Port 8184: 1000ms (very high latency)
- Port 8185: 3000ms (extreme latency)

**Best for:** Testing timeout configurations and understanding latency impacts

### `valid/ipv6.json`

**Use case:** IPv6 upstream testing  
**Routes:** 1 route on port 8186  
**Chaos:** None  
**Best for:** Verifying IPv6 support

## Invalid Configurations

These configurations demonstrate various validation errors. Useful for testing error handling and understanding configuration requirements.

### `invalid/duplicate_ports.json`

**Error:** Two routes using the same local port (8180)  
**Validates:** Port uniqueness checking

### `invalid/invalid_upstream_hostname.json`

**Error:** Using hostname "localhost" instead of IP address  
**Validates:** IP-only upstream requirement

### `invalid/invalid_port_range.json`

**Error:** Port 99999 exceeds valid range (1-65535)  
**Validates:** Port range validation

### `invalid/invalid_drop_rate.json`

**Error:** Drop rate 1.5 exceeds valid range (0.0-1.0)  
**Validates:** Drop rate bounds checking

### `invalid/invalid_negative_latency.json`

**Error:** Negative latency value (-100ms)  
**Validates:** Non-negative latency requirement

### `invalid/invalid_missing_upstream.json`

**Error:** Empty upstream field  
**Validates:** Required field checking

### `invalid/invalid_zero_port.json`

**Error:** Port 0 not allowed (requires static port assignment)  
**Validates:** Non-zero port requirement

## Testing with the `-test-server` Flag

You can use any of these configurations with the `-test-server` flag to automatically start HTTP test servers on all upstream targets:

```bash
./chaos-proxy -config examples/configs/valid/multiple_routes.json -test-server -verbose
```

This will start:

- Test HTTP servers on the upstream addresses (127.0.0.1:3000-3002)
- Proxy listeners on ports 8180-8182

Then test with:

```bash
curl http://127.0.0.1:8180/  # Stable service
curl http://127.0.0.1:8181/  # Light chaos (5% drop)
curl http://127.0.0.1:8182/  # Moderate chaos (15% drop, 500ms latency)
```
