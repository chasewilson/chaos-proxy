package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		wantLen     int
	}{
		{
			name: "valid config with single route",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "http://localhost:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				}
			]`,
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "valid config with multiple routes",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "http://localhost:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				},
				{
					"localPort": 8081,
					"upstream": "http://localhost:9091",
					"dropRate": 0.2,
					"latencyMs": 200
				}
			]`,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:        "empty config array",
			fileContent: `[]`,
			wantErr:     false,
			wantLen:     0,
		},
		{
			name:        "invalid JSON",
			fileContent: `{invalid json}`,
			wantErr:     true,
		},
		{
			name: "unknown fields should error",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "http://localhost:9090",
					"dropRate": 0.1,
					"latencyMs": 100,
					"unknownField": "value"
				}
			]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			if err := os.WriteFile(configPath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("failed to write test config file: %v", err)
			}

			config, err := LoadConfig(configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(config) != tt.wantLen {
				t.Errorf("LoadConfig() got %d routes, want %d", len(config), tt.wantLen)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("LoadConfig() expected error for nonexistent file, got nil")
	}
}

func TestLoadConfig_ValidFields(t *testing.T) {
	fileContent := `[
		{
			"localPort": 8080,
			"upstream": "http://localhost:9090",
			"dropRate": 0.5,
			"latencyMs": 250
		}
	]`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	if len(config) != 1 {
		t.Fatalf("expected 1 route, got %d", len(config))
	}

	route := config[0]
	if route.LocalPort != 8080 {
		t.Errorf("LocalPort = %d, want 8080", route.LocalPort)
	}
	if route.Upstream != "http://localhost:9090" {
		t.Errorf("Upstream = %s, want http://localhost:9090", route.Upstream)
	}
	if route.DropRate != 0.5 {
		t.Errorf("DropRate = %f, want 0.5", route.DropRate)
	}
	if route.LatencyMs != 250 {
		t.Errorf("LatencyMs = %d, want 250", route.LatencyMs)
	}
}
