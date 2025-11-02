package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func NewLogger(verbose, quiet bool) {
	var level slog.Level

	if verbose && quiet {
		fmt.Fprintln(os.Stderr, "warning: both -verbose and --quiet set. quiet mode takes precedence")
		verbose = false
	}

	switch {
	case verbose:
		level = slog.LevelDebug
	case quiet:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	options := slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stderr, &options)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
