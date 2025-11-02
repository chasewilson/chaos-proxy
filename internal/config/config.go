package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
)

type RouteConfig struct {
	LocalPort int     `json:"localPort"`
	Upstream  string  `json:"upstream"`
	DropRate  float64 `json:"dropRate"`
	LatencyMs int     `json:"latencyMs"`
}

// LoadConfig loads the route configuration from a JSON file.
func LoadConfig(configPath string) ([]RouteConfig, error) {
	configLogger := slog.With("file", configPath)
	file, err := os.Open(configPath)
	if err != nil {
		configLogger.Error("failed to open config file",
			"error", err,
			"hint", "check that the file exists and you have read permissions")
		return nil, fmt.Errorf("cannot open config file %q: %w", configPath, err)
	}
	defer file.Close()

	var config []RouteConfig
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		configLogger.Error("invalid JSON in config file",
			"error", err,
			"hint", "verify JSON syntax is valid (check for missing commas, quotes, brackets)")
		return nil, fmt.Errorf("invalid JSON in config file %q: %w", configPath, err)
	}

	if err := validateConfig(config, configLogger); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(routes []RouteConfig, configLogger *slog.Logger) error {
	if len(routes) == 0 {
		configLogger.Error("empty route configuration", "hint", "config file must contain at least one route")
		return fmt.Errorf("validation failed: empty route configuration")
	}

	portMap := make(map[int]struct{})
	hasErrors := false

	for i, route := range routes {
		if err := validateRouteConfig(route, i, configLogger); err != nil {
			hasErrors = true
		}

		if _, exists := portMap[route.LocalPort]; exists {
			configLogger.Error("duplicate local port detected",
				"port", route.LocalPort,
				"route_index", i,
				"hint", fmt.Sprintf("each route must use a unique localPort. Port %d is already used by another route", route.LocalPort))
			hasErrors = true
		} else {
			portMap[route.LocalPort] = struct{}{}
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed: see error messages above for details")
	}

	return nil
}

func validateRouteConfig(config RouteConfig, routeIndex int, configLogger *slog.Logger) error {
	hasErrors := false
	routeLogger := configLogger.With("route_index", routeIndex)

	// Validate local port - 0 isn't allowed. Require static port assignment.
	if config.LocalPort <= 0 || config.LocalPort > 65535 {
		routeLogger.Error("invalid local port",
			"port", config.LocalPort,
			"valid_range", "1-65535",
			"hint", fmt.Sprintf("localPort must be between 1 and 65535, got %d", config.LocalPort))
		hasErrors = true
	}

	if config.Upstream == "" {
		routeLogger.Error("upstream field is empty",
			"hint", "upstream must be in format 'ip:port' (e.g., '127.0.0.1:9090')")
		hasErrors = true
	} else {
		host, port, err := net.SplitHostPort(config.Upstream)
		if err != nil {
			routeLogger.Error("invalid upstream format",
				"upstream", config.Upstream,
				"error", err,
				"hint", "upstream must be in format 'ip:port' (e.g., '127.0.0.1:9090' or '[::1]:9090' for IPv6)")
			hasErrors = true
		} else {
			if net.ParseIP(host) == nil {
				routeLogger.Error("upstream host is not a valid IP address",
					"upstream", config.Upstream,
					"host", host,
					"hint", fmt.Sprintf("host must be an IP address (e.g., '127.0.0.1' or '[::1]'), not a hostname. Got %q", host))
				hasErrors = true
			}

			portNum, err := strconv.Atoi(port)
			if err != nil {
				routeLogger.Error("upstream port is not a number",
					"upstream", config.Upstream,
					"port", port,
					"error", err,
					"hint", fmt.Sprintf("port must be a number between 1-65535, got %q", port))
				hasErrors = true
			} else if portNum <= 0 || portNum > 65535 {
				routeLogger.Error("upstream port out of valid range",
					"upstream", config.Upstream,
					"port", portNum,
					"valid_range", "1-65535",
					"hint", fmt.Sprintf("port must be between 1 and 65535, got %d", portNum))
				hasErrors = true
			}
		}
	}

	if config.DropRate < 0.0 || config.DropRate > 1.0 {
		routeLogger.Error("invalid drop rate",
			"drop_rate", config.DropRate,
			"valid_range", "0.0-1.0",
			"hint", fmt.Sprintf("dropRate must be between 0.0 and 1.0 (probability), got %.2f", config.DropRate))
		hasErrors = true
	}

	if config.LatencyMs < 0 {
		routeLogger.Error("invalid latency",
			"latency_ms", config.LatencyMs,
			"valid_range", ">= 0",
			"hint", fmt.Sprintf("latencyMs must be >= 0 (milliseconds), got %d", config.LatencyMs))
		hasErrors = true
	}

	if hasErrors {
		return fmt.Errorf("route[%d] validation failed", routeIndex)
	}

	return nil
}
