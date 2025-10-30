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

func LoadConfig(configPath string) ([]RouteConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var config []RouteConfig
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return config, nil
}
