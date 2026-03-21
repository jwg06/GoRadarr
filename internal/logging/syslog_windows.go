//go:build windows

package logging

import "log/slog"

// setupSyslog is a no-op on Windows; caller will fall back to stderr.
func setupSyslog(_ *slog.HandlerOptions, _ SyslogConfig) *slog.Logger { return nil }
