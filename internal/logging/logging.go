package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Setup creates a configured slog.Logger.
//
// target: "stderr" | "stdout" | "file" | "syslog"
// level:  "debug" | "info" | "warn" | "error"
// file:   path used when target == "file"
func Setup(level, target, file string) *slog.Logger {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: lvl == slog.LevelDebug,
	}

	switch strings.ToLower(target) {
	case "syslog":
		if l := setupSyslog(opts); l != nil {
			return l
		}
		// syslog unavailable — fall through to stderr
		fallthrough
	case "file":
		if file != "" {
			if err := os.MkdirAll(filepath.Dir(file), 0o755); err == nil {
				f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				if err == nil {
					return slog.New(newMultiHandler(opts, f))
				}
			}
		}
		fallthrough
	case "stdout":
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	default: // "stderr" and anything else
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// newMultiHandler returns a JSON handler writing to w.
func newMultiHandler(opts *slog.HandlerOptions, w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, opts)
}
