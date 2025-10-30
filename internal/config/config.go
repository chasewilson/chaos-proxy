package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type RouteConfig struct {
	LocalPort int     `json:"localPort"`
	Upstream  string  `json:"upstream"`
	DropRate  float64 `json:"dropRate"`
	LatencyMs int     `json:"latencyMs"`
}

func LoadConfig(configPath string) ([]RouteConfig, []error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, []error{fmt.Errorf("error opening file: %w", err)}
	}
	defer file.Close()

	var config []RouteConfig
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return nil, []error{fmt.Errorf("invalid JSON: %w", err)}
	}

	validationErrors := validateConfig(config)
	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	return config, nil
}

func validateConfig(routes []RouteConfig) []error {
	portMap := make(map[int]struct{})
	var errs []error

	for i, route := range routes {
		errs = append(errs, validateRouteConfig(route, i)...)

		if _, exists := portMap[route.LocalPort]; exists {
			errs = append(errs, fmt.Errorf("cannot use duplicate local port %d (route index: %d)", route.LocalPort, i))
		} else {
			portMap[route.LocalPort] = struct{}{}
		}
	}

	return errs
}

func validateRouteConfig(config RouteConfig, routeIndex int) []error {
	var errs []error
	// Validate local port - 0 isn't allowed. Require static port assignment.
	if config.LocalPort <= 0 || config.LocalPort > 65535 {
		errs = append(errs, fmt.Errorf("invalid local port: %d (route index: %d)", config.LocalPort, routeIndex))
	}
	if config.Upstream == "" {
		errs = append(errs, fmt.Errorf("'upstream' cannot be empty (route index: %d)", routeIndex))
	}
	if config.DropRate < 0.0 || config.DropRate > 1.0 {
		errs = append(errs, fmt.Errorf("invalid drop rate: %f (route index: %d)", config.DropRate, routeIndex))
	}
	if config.LatencyMs < 0 {
		errs = append(errs, fmt.Errorf("invalid latency: %d (route index: %d)", config.LatencyMs, routeIndex))
	}
	return errs
}

// TODO: Add error sorting return functionality to improve user experience
