//go:build !windows

package logging

import (
	"log/slog"
	"log/syslog"
)

// setupSyslog wires slog to the local syslog daemon.
// The syslog writer implements io.Writer, so we wrap it in a JSON handler;
// each log line arrives as a single structured JSON object in the syslog entry.
func setupSyslog(opts *slog.HandlerOptions) *slog.Logger {
	w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "goradarr")
	if err != nil {
		return nil
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}
