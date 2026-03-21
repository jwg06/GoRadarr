package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestSetupStderr(t *testing.T) {
	l := Setup("info", "stderr", "", SyslogConfig{})
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestSetupStdout(t *testing.T) {
	l := Setup("debug", "stdout", "", SyslogConfig{})
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestSetupFile(t *testing.T) {
	tmp := t.TempDir() + "/test.log"
	l := Setup("info", "file", tmp, SyslogConfig{})
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	l.Info("hello from file logger")
}

func TestSetupSyslogBadAddress(t *testing.T) {
	// An unreachable address should fall back gracefully (no panic).
	l := Setup("info", "syslog", "", SyslogConfig{Address: "127.0.0.1", Port: 1, Network: "udp"})
	if l == nil {
		t.Fatal("expected non-nil fallback logger when syslog is unreachable")
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}
	for _, c := range cases {
		if got := parseLevel(c.in); got != c.want {
			t.Errorf("parseLevel(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}

func TestNewMultiHandlerJSON(t *testing.T) {
	var buf bytes.Buffer
	h := newMultiHandler(&slog.HandlerOptions{Level: slog.LevelInfo}, &buf)
	l := slog.New(h)
	l.Info("test message", "key", "value")
	out := buf.String()
	if !strings.Contains(out, "test message") {
		t.Errorf("expected log output to contain 'test message', got: %s", out)
	}
	if !strings.Contains(out, `"key"`) {
		t.Errorf("expected JSON key field, got: %s", out)
	}
}
