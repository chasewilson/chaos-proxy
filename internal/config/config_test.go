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
					"upstream": "127.0.0.1:9090",
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
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				},
				{
					"localPort": 8081,
					"upstream": "192.168.1.1:9091",
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
					"upstream": "127.0.0.1:9090",
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

			config, errs := LoadConfig(configPath)

			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("LoadConfig() errors = %v, wantErr %v", errs, tt.wantErr)
				return
			}

			if !tt.wantErr && len(config) != tt.wantLen {
				t.Errorf("LoadConfig() got %d routes, want %d", len(config), tt.wantLen)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, errs := LoadConfig("/nonexistent/path/config.json")
	if len(errs) == 0 {
		t.Error("LoadConfig() expected error for nonexistent file, got nil")
	}
}

func TestLoadConfig_ValidFields(t *testing.T) {
	fileContent := `[
		{
			"localPort": 8080,
			"upstream": "127.0.0.1:9090",
			"dropRate": 0.5,
			"latencyMs": 250
		}
	]`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	config, errs := LoadConfig(configPath)
	if len(errs) > 0 {
		t.Fatalf("LoadConfig() unexpected error: %v", errs)
	}

	if len(config) != 1 {
		t.Fatalf("expected 1 route, got %d", len(config))
	}

	route := config[0]
	if route.LocalPort != 8080 {
		t.Errorf("LocalPort = %d, want 8080", route.LocalPort)
	}
	if route.Upstream != "127.0.0.1:9090" {
		t.Errorf("Upstream = %s, want 127.0.0.1:9090", route.Upstream)
	}
	if route.DropRate != 0.5 {
		t.Errorf("DropRate = %f, want 0.5", route.DropRate)
	}
	if route.LatencyMs != 250 {
		t.Errorf("LatencyMs = %d, want 250", route.LatencyMs)
	}
}

func TestLoadConfig_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid port - zero",
			fileContent: `[
				{
					"localPort": 0,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				}
			]`,
			wantErr:     true,
			errContains: "invalid local port",
		},
		{
			name: "invalid port - negative",
			fileContent: `[
				{
					"localPort": -1,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				}
			]`,
			wantErr:     true,
			errContains: "invalid local port",
		},
		{
			name: "invalid port - too large",
			fileContent: `[
				{
					"localPort": 65536,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				}
			]`,
			wantErr:     true,
			errContains: "invalid local port",
		},
		{
			name: "empty upstream",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "",
					"dropRate": 0.1,
					"latencyMs": 100
				}
			]`,
			wantErr:     true,
			errContains: "upstream",
		},
		{
			name: "invalid drop rate - negative",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9090",
					"dropRate": -0.1,
					"latencyMs": 100
				}
			]`,
			wantErr:     true,
			errContains: "invalid drop rate",
		},
		{
			name: "invalid drop rate - too large",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9090",
					"dropRate": 1.5,
					"latencyMs": 100
				}
			]`,
			wantErr:     true,
			errContains: "invalid drop rate",
		},
		{
			name: "invalid latency - negative",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": -100
				}
			]`,
			wantErr:     true,
			errContains: "invalid latency",
		},
		{
			name: "upstream with hostname instead of IP",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "localhost:9090",
					"dropRate": 0.0,
					"latencyMs": 0
				}
			]`,
			wantErr:     true,
			errContains: "host must be a valid IP address",
		},
		{
			name: "upstream with URL scheme",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "http://127.0.0.1:9090",
					"dropRate": 0.0,
					"latencyMs": 0
				}
			]`,
			wantErr:     true,
			errContains: "invalid upstream format",
		},
		{
			name: "upstream missing port",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1",
					"dropRate": 0.0,
					"latencyMs": 0
				}
			]`,
			wantErr:     true,
			errContains: "invalid upstream format",
		},
		{
			name: "upstream with invalid port",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:99999",
					"dropRate": 0.0,
					"latencyMs": 0
				}
			]`,
			wantErr:     true,
			errContains: "invalid upstream port",
		},
		{
			name: "valid edge case - port 1",
			fileContent: `[
				{
					"localPort": 1,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.0,
					"latencyMs": 0
				}
			]`,
			wantErr: false,
		},
		{
			name: "valid edge case - port 65535",
			fileContent: `[
				{
					"localPort": 65535,
					"upstream": "127.0.0.1:9090",
					"dropRate": 1.0,
					"latencyMs": 0
				}
			]`,
			wantErr: false,
		},
		{
			name: "valid IPv6 address",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "[::1]:9090",
					"dropRate": 0.0,
					"latencyMs": 0
				}
			]`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			if err := os.WriteFile(configPath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("failed to write test config file: %v", err)
			}

			_, errs := LoadConfig(configPath)

			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("LoadConfig() errors = %v, wantErr %v", errs, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if len(errs) == 0 {
					t.Errorf("LoadConfig() expected errors, got none")
				} else {
					foundMatch := false
					for _, err := range errs {
						if contains(err.Error(), tt.errContains) {
							foundMatch = true
							break
						}
					}
					if !foundMatch {
						t.Errorf("LoadConfig() errors = %v, want error containing %q", errs, tt.errContains)
					}
				}
			}
		})
	}
}

func TestLoadConfig_DuplicatePorts(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		errContains string
	}{
		{
			name: "duplicate ports in different routes",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				},
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9091",
					"dropRate": 0.2,
					"latencyMs": 200
				}
			]`,
			wantErr:     true,
			errContains: "cannot use duplicate local port",
		},
		{
			name: "no duplicate ports",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				},
				{
					"localPort": 8081,
					"upstream": "127.0.0.1:9091",
					"dropRate": 0.2,
					"latencyMs": 200
				}
			]`,
			wantErr: false,
		},
		{
			name: "three routes with duplicate port",
			fileContent: `[
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9090",
					"dropRate": 0.1,
					"latencyMs": 100
				},
				{
					"localPort": 8081,
					"upstream": "127.0.0.1:9091",
					"dropRate": 0.2,
					"latencyMs": 200
				},
				{
					"localPort": 8080,
					"upstream": "127.0.0.1:9092",
					"dropRate": 0.3,
					"latencyMs": 300
				}
			]`,
			wantErr:     true,
			errContains: "cannot use duplicate local port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			if err := os.WriteFile(configPath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("failed to write test config file: %v", err)
			}

			_, errs := LoadConfig(configPath)

			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("LoadConfig() errors = %v, wantErr %v", errs, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if len(errs) == 0 {
					t.Errorf("LoadConfig() expected errors, got none")
				} else {
					foundMatch := false
					for _, err := range errs {
						if contains(err.Error(), tt.errContains) {
							foundMatch = true
							break
						}
					}
					if !foundMatch {
						t.Errorf("LoadConfig() errors = %v, want error containing %q", errs, tt.errContains)
					}
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		routes      []RouteConfig
		wantErrLen  int
		errContains []string
	}{
		{
			name: "valid routes",
			routes: []RouteConfig{
				{
					LocalPort: 8080,
					Upstream:  "127.0.0.1:9090",
					DropRate:  0.1,
					LatencyMs: 100,
				},
				{
					LocalPort: 8081,
					Upstream:  "192.168.1.1:9091",
					DropRate:  0.2,
					LatencyMs: 200,
				},
			},
			wantErrLen: 0,
		},
		{
			name: "duplicate ports",
			routes: []RouteConfig{
				{
					LocalPort: 8080,
					Upstream:  "127.0.0.1:9090",
					DropRate:  0.1,
					LatencyMs: 100,
				},
				{
					LocalPort: 8080,
					Upstream:  "127.0.0.1:9091",
					DropRate:  0.2,
					LatencyMs: 200,
				},
			},
			wantErrLen:  1,
			errContains: []string{"cannot use duplicate local port"},
		},
		{
			name: "invalid route and duplicate port",
			routes: []RouteConfig{
				{
					LocalPort: 8080,
					Upstream:  "",
					DropRate:  0.1,
					LatencyMs: 100,
				},
				{
					LocalPort: 8080,
					Upstream:  "127.0.0.1:9091",
					DropRate:  0.2,
					LatencyMs: 200,
				},
			},
			wantErrLen:  2,
			errContains: []string{"upstream", "cannot use duplicate local port"},
		},
		{
			name: "multiple invalid routes",
			routes: []RouteConfig{
				{
					LocalPort: 0,
					Upstream:  "",
					DropRate:  -0.1,
					LatencyMs: -100,
				},
				{
					LocalPort: 70000,
					Upstream:  "localhost:9091",
					DropRate:  1.5,
					LatencyMs: 200,
				},
			},
			wantErrLen:  7,
			errContains: []string{"invalid local port", "upstream", "invalid drop rate", "invalid latency", "host must be a valid IP address"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateConfig(tt.routes)

			if len(errs) != tt.wantErrLen {
				t.Errorf("validateConfig() got %d errors, want %d. Errors: %v", len(errs), tt.wantErrLen, errs)
				return
			}

			for _, expectedErr := range tt.errContains {
				foundMatch := false
				for _, err := range errs {
					if contains(err.Error(), expectedErr) {
						foundMatch = true
						break
					}
				}
				if !foundMatch {
					t.Errorf("validateConfig() errors = %v, want error containing %q", errs, expectedErr)
				}
			}
		})
	}
}

func TestValidateRouteConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      RouteConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr: false,
		},
		{
			name: "valid config - minimum values",
			config: RouteConfig{
				LocalPort: 1,
				Upstream:  "127.0.0.1:9090",
				DropRate:  0.0,
				LatencyMs: 0,
			},
			wantErr: false,
		},
		{
			name: "valid config - maximum values",
			config: RouteConfig{
				LocalPort: 65535,
				Upstream:  "192.168.1.100:65535",
				DropRate:  1.0,
				LatencyMs: 999999,
			},
			wantErr: false,
		},
		{
			name: "valid config - IPv6",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "[::1]:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr: false,
		},
		{
			name: "valid config - IPv6 full address",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "[2001:db8::1]:8080",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr: false,
		},
		{
			name: "invalid port - zero",
			config: RouteConfig{
				LocalPort: 0,
				Upstream:  "127.0.0.1:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid local port",
		},
		{
			name: "invalid port - negative",
			config: RouteConfig{
				LocalPort: -1,
				Upstream:  "127.0.0.1:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid local port",
		},
		{
			name: "invalid port - too large",
			config: RouteConfig{
				LocalPort: 65536,
				Upstream:  "127.0.0.1:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid local port",
		},
		{
			name: "empty upstream",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "upstream",
		},
		{
			name: "invalid drop rate - negative",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:9090",
				DropRate:  -0.1,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid drop rate",
		},
		{
			name: "invalid drop rate - too large",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:9090",
				DropRate:  1.1,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid drop rate",
		},
		{
			name: "invalid latency - negative",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:9090",
				DropRate:  0.5,
				LatencyMs: -1,
			},
			wantErr:     true,
			errContains: "invalid latency",
		},
		{
			name: "upstream with hostname",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "localhost:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "host must be a valid IP address",
		},
		{
			name: "upstream with domain name",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "example.com:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "host must be a valid IP address",
		},
		{
			name: "upstream with URL scheme",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "http://127.0.0.1:9090",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid upstream format",
		},
		{
			name: "upstream missing port",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid upstream format",
		},
		{
			name: "upstream with invalid port - too high",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:99999",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid upstream port",
		},
		{
			name: "upstream with invalid port - zero",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:0",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid upstream port",
		},
		{
			name: "upstream with invalid port - negative",
			config: RouteConfig{
				LocalPort: 8080,
				Upstream:  "127.0.0.1:-1",
				DropRate:  0.5,
				LatencyMs: 100,
			},
			wantErr:     true,
			errContains: "invalid upstream port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateRouteConfig(tt.config, 0)

			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("validateRouteConfig() errors = %v, wantErr %v", errs, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if len(errs) == 0 {
					t.Errorf("validateRouteConfig() expected errors, got none")
				} else {
					foundMatch := false
					for _, err := range errs {
						if contains(err.Error(), tt.errContains) {
							foundMatch = true
							break
						}
					}
					if !foundMatch {
						t.Errorf("validateRouteConfig() errors = %v, want error containing %q", errs, tt.errContains)
					}
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
