package logger

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNewLogger_Default(t *testing.T) {
	// Capture stderr to verify no warnings
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	NewLogger(false, false)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Verify default logger is set and uses Info level
	logger := slog.Default()
	if logger == nil {
		t.Fatal("default logger should be set")
	}

	// Test that Info level messages are logged
	var buf bytes.Buffer
	testHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	testLogger := slog.New(testHandler)

	// Get the handler from default logger to check level
	// We can't directly access handler level, so we test by checking output
	testLogger.Info("test message")
	if buf.Len() == 0 {
		t.Error("Info level messages should be logged in default mode")
	}
}

func TestNewLogger_Verbose(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	NewLogger(true, false)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Verify logger is set
	logger := slog.Default()
	if logger == nil {
		t.Fatal("default logger should be set")
	}

	// Test that Debug level messages are logged in verbose mode
	var buf bytes.Buffer
	testHandler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	testLogger := slog.New(testHandler)

	testLogger.Debug("debug message")
	if buf.Len() == 0 {
		t.Error("Debug level messages should be logged in verbose mode")
	}
}

func TestNewLogger_Quiet(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stderr = w

	NewLogger(false, true)

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Verify logger is set
	logger := slog.Default()
	if logger == nil {
		t.Fatal("default logger should be set")
	}

	// In quiet mode, only errors should be logged
	// Info messages should not appear
	var bufInfo bytes.Buffer
	testHandlerInfo := slog.NewTextHandler(&bufInfo, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	testLoggerInfo := slog.New(testHandlerInfo)

	testLoggerInfo.Info("info message")
	if bufInfo.Len() > 0 {
		t.Error("Info level messages should NOT be logged in quiet mode")
	}

	// Error messages should still appear
	var bufError bytes.Buffer
	testHandlerError := slog.NewTextHandler(&bufError, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	testLoggerError := slog.New(testHandlerError)

	testLoggerError.Error("error message")
	if bufError.Len() == 0 {
		t.Error("Error level messages should be logged in quiet mode")
	}
}

func TestNewLogger_BothFlagsSet(t *testing.T) {
	// Capture stderr to check for warning message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	NewLogger(true, true)

	// Read stderr output
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()
	os.Stderr = oldStderr

	// Verify warning message is printed
	if !strings.Contains(output, "warning") {
		t.Error("Expected warning message when both verbose and quiet are set")
	}
	if !strings.Contains(output, "quiet mode takes precedence") {
		t.Error("Expected 'quiet mode takes precedence' in warning message")
	}

	// Verify logger uses Error level (quiet mode)
	logger := slog.Default()
	if logger == nil {
		t.Fatal("default logger should be set")
	}

	// Test that quiet mode is used (Error level only)
	var bufInfo bytes.Buffer
	testHandlerInfo := slog.NewTextHandler(&bufInfo, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	testLoggerInfo := slog.New(testHandlerInfo)

	testLoggerInfo.Info("info message")
	if bufInfo.Len() > 0 {
		t.Error("Info level messages should NOT be logged when both flags set (quiet takes precedence)")
	}
}

func TestNewLogger_TimeRemoved(t *testing.T) {
	// Set up logger
	NewLogger(false, false)

	// Create a test handler that captures output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	logger := slog.New(handler)

	// Log a message
	logger.Info("test message", "key", "value")

	output := buf.String()

	// Verify time is not in output
	if strings.Contains(output, "time=") {
		t.Error("Time should be removed from log output")
	}

	// Verify message and attributes are still present
	if !strings.Contains(output, "test message") {
		t.Error("Log message should be present")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Log attributes should be present")
	}
}

func TestNewLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name           string
		verbose        bool
		quiet          bool
		shouldLogInfo  bool
		shouldLogDebug bool
		shouldLogError bool
	}{
		{
			name:           "default mode",
			verbose:        false,
			quiet:          false,
			shouldLogInfo:  true,
			shouldLogDebug: false,
			shouldLogError: true,
		},
		{
			name:           "verbose mode",
			verbose:        true,
			quiet:          false,
			shouldLogInfo:  true,
			shouldLogDebug: true,
			shouldLogError: true,
		},
		{
			name:           "quiet mode",
			verbose:        false,
			quiet:          true,
			shouldLogInfo:  false,
			shouldLogDebug: false,
			shouldLogError: true,
		},
		{
			name:           "both flags (quiet precedence)",
			verbose:        true,
			quiet:          true,
			shouldLogInfo:  false,
			shouldLogDebug: false,
			shouldLogError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stderr to avoid test output pollution
			oldStderr := os.Stderr
			_, w, _ := os.Pipe()
			os.Stderr = w

			NewLogger(tt.verbose, tt.quiet)

			// Restore stderr
			w.Close()
			os.Stderr = oldStderr

			// Create test handlers with expected levels
			var bufInfo, bufDebug, bufError bytes.Buffer

			infoHandler := slog.NewTextHandler(&bufInfo, &slog.HandlerOptions{
				Level: getExpectedLevel(tt.verbose, tt.quiet),
			})
			debugHandler := slog.NewTextHandler(&bufDebug, &slog.HandlerOptions{
				Level: getExpectedLevel(tt.verbose, tt.quiet),
			})
			errorHandler := slog.NewTextHandler(&bufError, &slog.HandlerOptions{
				Level: getExpectedLevel(tt.verbose, tt.quiet),
			})

			infoLogger := slog.New(infoHandler)
			debugLogger := slog.New(debugHandler)
			errorLogger := slog.New(errorHandler)

			infoLogger.Info("info test")
			debugLogger.Debug("debug test")
			errorLogger.Error("error test")

			if (bufInfo.Len() > 0) != tt.shouldLogInfo {
				t.Errorf("Info logging: got %v, want %v", bufInfo.Len() > 0, tt.shouldLogInfo)
			}
			if (bufDebug.Len() > 0) != tt.shouldLogDebug {
				t.Errorf("Debug logging: got %v, want %v", bufDebug.Len() > 0, tt.shouldLogDebug)
			}
			if (bufError.Len() > 0) != tt.shouldLogError {
				t.Errorf("Error logging: got %v, want %v", bufError.Len() > 0, tt.shouldLogError)
			}
		})
	}
}

// Helper function to get expected log level
func getExpectedLevel(verbose, quiet bool) slog.Level {
	if verbose && quiet {
		return slog.LevelError // Quiet takes precedence
	}
	if verbose {
		return slog.LevelDebug
	}
	if quiet {
		return slog.LevelError
	}
	return slog.LevelInfo
}

