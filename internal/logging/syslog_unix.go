//go:build !windows

package logging

import (
	"fmt"
	"log/slog"
	"log/syslog"
)

// setupSyslog connects to a syslog target.
// When sc.Address is set, it dials the remote host over UDP/TCP.
// When sc.Address is empty, it connects to the local syslog daemon.
// Each log record is written as a single structured JSON line.
func setupSyslog(opts *slog.HandlerOptions, sc SyslogConfig) *slog.Logger {
	priority := syslog.LOG_INFO | syslog.LOG_DAEMON

	var w *syslog.Writer
	var err error

	if sc.Address != "" {
		network := sc.Network
		if network == "" {
			network = "udp"
		}
		port := sc.Port
		if port == 0 {
			port = 514
		}
		raddr := fmt.Sprintf("%s:%d", sc.Address, port)
		w, err = syslog.Dial(network, raddr, priority, "goradarr")
	} else {
		w, err = syslog.New(priority, "goradarr")
	}

	if err != nil {
		return nil
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}
