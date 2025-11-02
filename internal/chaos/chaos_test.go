package chaos

import (
	"math"
	"testing"
	"time"
)

func TestNewCurse_DropRate(t *testing.T) {
	tests := []struct {
		name     string
		dropRate float64
		wantDrop bool
	}{
		{
			name:     "zero drop rate",
			dropRate: 0.0,
			wantDrop: false,
		},
		{
			name:     "low drop rate",
			dropRate: 0.1,
			wantDrop: false, // Not deterministic, but test that it can be false
		},
		{
			name:     "one hundred percent drop rate",
			dropRate: 1.0,
			wantDrop: true, // Should always drop
		},
		{
			name:     "negative drop rate",
			dropRate: -0.1,
			wantDrop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ritual := Ritual{
				DropRate:  tt.dropRate,
				LatencyMs: 0,
			}

			// For deterministic cases (0.0 and 1.0), test directly
			if tt.dropRate == 0.0 {
				curse := NewCurse(ritual)
				if curse.DropConnections != false {
					t.Errorf("NewCurse() DropConnections = %v, want false for dropRate 0.0", curse.DropConnections)
				}
			} else if tt.dropRate == 1.0 {
				curse := NewCurse(ritual)
				if curse.DropConnections != true {
					t.Errorf("NewCurse() DropConnections = %v, want true for dropRate 1.0", curse.DropConnections)
				}
			} else {
				// For non-deterministic cases, just verify it doesn't crash
				curse := NewCurse(ritual)
				if curse.DropConnections && tt.dropRate <= 0 {
					t.Errorf("NewCurse() DropConnections = true with dropRate %f, should be false", tt.dropRate)
				}
			}
		})
	}
}

func TestNewCurse_DropRateProbability(t *testing.T) {
	// Statistical test: With dropRate 0.5, we should get approximately 50% drops
	dropRate := 0.5
	iterations := 1000
	drops := 0

	for i := 0; i < iterations; i++ {
		ritual := Ritual{
			DropRate:  dropRate,
			LatencyMs: 0,
		}
		curse := NewCurse(ritual)
		if curse.DropConnections {
			drops++
		}
	}

	actualRate := float64(drops) / float64(iterations)
	// Allow 10% deviation from expected 50%
	expectedRate := 0.5
	tolerance := 0.1

	if math.Abs(actualRate-expectedRate) > tolerance {
		t.Errorf("Drop rate probability test: got %.2f%%, want %.2f%% ± %.2f%%", actualRate*100, expectedRate*100, tolerance*100)
	}
}

func TestNewCurse_LatencyMs(t *testing.T) {
	tests := []struct {
		name      string
		latencyMs int
		wantDelay time.Duration
	}{
		{
			name:      "zero latency",
			latencyMs: 0,
			wantDelay: 0,
		},
		{
			name:      "small latency",
			latencyMs: 50,
			wantDelay: 50 * time.Millisecond,
		},
		{
			name:      "medium latency",
			latencyMs: 200,
			wantDelay: 200 * time.Millisecond,
		},
		{
			name:      "large latency",
			latencyMs: 1000,
			wantDelay: 1000 * time.Millisecond,
		},
		{
			name:      "negative latency",
			latencyMs: -100,
			wantDelay: 0, // Should not set delay for negative values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ritual := Ritual{
				DropRate:  0.0,
				LatencyMs: tt.latencyMs,
			}

			curse := NewCurse(ritual)

			if tt.latencyMs > 0 {
				if curse.StartDelay != tt.wantDelay {
					t.Errorf("NewCurse() StartDelay = %v, want %v", curse.StartDelay, tt.wantDelay)
				}
			} else {
				if curse.StartDelay != 0 {
					t.Errorf("NewCurse() StartDelay = %v, want 0 for latencyMs %d", curse.StartDelay, tt.latencyMs)
				}
			}
		})
	}
}

func TestNewCurse_Combined(t *testing.T) {
	tests := []struct {
		name      string
		dropRate  float64
		latencyMs int
	}{
		{
			name:      "both zero",
			dropRate:  0.0,
			latencyMs: 0,
		},
		{
			name:      "drop rate only",
			dropRate:  0.3,
			latencyMs: 0,
		},
		{
			name:      "latency only",
			dropRate:  0.0,
			latencyMs: 100,
		},
		{
			name:      "both set",
			dropRate:  0.5,
			latencyMs: 200,
		},
		{
			name:      "high drop rate with latency",
			dropRate:  0.9,
			latencyMs: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ritual := Ritual{
				DropRate:  tt.dropRate,
				LatencyMs: tt.latencyMs,
			}

			curse := NewCurse(ritual)

			// Verify drop connection is set correctly
			if tt.dropRate == 0.0 {
				if curse.DropConnections {
					t.Errorf("NewCurse() DropConnections = true, want false for dropRate 0.0")
				}
			} else if tt.dropRate == 1.0 {
				if !curse.DropConnections {
					t.Errorf("NewCurse() DropConnections = false, want true for dropRate 1.0")
				}
			}

			// Verify latency is set correctly
			if tt.latencyMs > 0 {
				expectedDelay := time.Duration(tt.latencyMs) * time.Millisecond
				if curse.StartDelay != expectedDelay {
					t.Errorf("NewCurse() StartDelay = %v, want %v", curse.StartDelay, expectedDelay)
				}
			} else {
				if curse.StartDelay != 0 {
					t.Errorf("NewCurse() StartDelay = %v, want 0 for latencyMs %d", curse.StartDelay, tt.latencyMs)
				}
			}
		})
	}
}

func TestNewCurse_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		dropRate  float64
		latencyMs int
	}{
		{
			name:      "minimum valid values",
			dropRate:  0.0,
			latencyMs: 0,
		},
		{
			name:      "maximum drop rate",
			dropRate:  1.0,
			latencyMs: 0,
		},
		{
			name:      "very small drop rate",
			dropRate:  0.0001,
			latencyMs: 0,
		},
		{
			name:      "very small latency",
			dropRate:  0.0,
			latencyMs: 1,
		},
		{
			name:      "very large latency",
			dropRate:  0.0,
			latencyMs: 3600000, // 1 hour in milliseconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ritual := Ritual{
				DropRate:  tt.dropRate,
				LatencyMs: tt.latencyMs,
			}

			curse := NewCurse(ritual)

			// Verify it doesn't panic and returns valid curse
			if curse.StartDelay < 0 {
				t.Errorf("NewCurse() StartDelay = %v, want non-negative", curse.StartDelay)
			}

			if tt.latencyMs > 0 {
				expectedDelay := time.Duration(tt.latencyMs) * time.Millisecond
				if curse.StartDelay != expectedDelay {
					t.Errorf("NewCurse() StartDelay = %v, want %v", curse.StartDelay, expectedDelay)
				}
			}
		})
	}
}

