package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// SyslogConfig carries remote syslog dial parameters.
type SyslogConfig struct {
	Address string // host or IP; empty = local syslog daemon
	Port    int    // default 514
	Network string // "udp" (default) | "tcp" | "unix"
}

// Setup creates a configured slog.Logger that writes to the requested target
// and also feeds the in-process ring buffer for the /api/v1/log endpoint.
//
// target: "stderr" | "stdout" | "file" | "syslog"
// level:  "debug" | "info" | "warn" | "error"
// file:   path used when target == "file"
func Setup(level, target, file string, sc SyslogConfig) *slog.Logger {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: lvl == slog.LevelDebug,
	}
	primary := buildPrimaryHandler(opts, target, file, sc)
	ring := RingHandler(opts)
	return slog.New(&teeHandler{primary: primary, ring: ring})
}

func buildPrimaryHandler(opts *slog.HandlerOptions, target, file string, sc SyslogConfig) slog.Handler {
	switch strings.ToLower(target) {
	case "syslog":
		if l := setupSyslog(opts, sc); l != nil {
			return l.Handler()
		}
		// syslog unavailable — fall through to stderr
		fallthrough
	case "file":
		if file != "" {
			if err := os.MkdirAll(filepath.Dir(file), 0o755); err == nil {
				f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
				if err == nil {
					return newMultiHandler(opts, f)
				}
			}
		}
		fallthrough
	case "stdout":
		return slog.NewTextHandler(os.Stdout, opts)
	default: // "stderr" and anything else
		return slog.NewTextHandler(os.Stderr, opts)
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
