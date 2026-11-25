// ABOUTME: Tests for New, ParseLevel, and Options defaults.
// ABOUTME: Verifies JSON/text format, service field, and writer override.

package hlog

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"  Debug ", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"FATAL", slog.LevelInfo},
	}
	for _, tt := range tests {
		if got := ParseLevel(tt.input); got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Format: "json", Service: "test-svc", Writer: &buf})
	logger.Info("hello")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON, got: %s", buf.String())
	}
	if m["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", m["msg"])
	}
	if m["service"] != "test-svc" {
		t.Errorf("service = %v, want test-svc", m["service"])
	}
}

func TestNewTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Format: "text", Service: "test-svc", Writer: &buf})
	logger.Info("hello")

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("hello")) {
		t.Errorf("expected 'hello' in output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("service=test-svc")) {
		t.Errorf("expected 'service=test-svc' in output: %s", out)
	}
}

func TestNewNoService(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Writer: &buf})
	logger.Info("hello")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON, got: %s", buf.String())
	}
	if _, ok := m["service"]; ok {
		t.Errorf("service should not be present when Options.Service is empty")
	}
}

func TestNewDefaultWriter(t *testing.T) {
	const message = "default writer probe"
	if os.Getenv("HLOG_TEST_DEFAULT_WRITER") == "1" {
		New(Options{}).Error(message)
		return
	}

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNewDefaultWriter$")
	cmd.Env = append(os.Environ(), "HLOG_TEST_DEFAULT_WRITER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("logging subprocess: %v: %s", err, stderr.String())
	}
	var entry struct {
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(stderr.Bytes(), &entry); err != nil {
		t.Fatalf("expected a JSON record on stderr: %v", err)
	}
	if entry.Message != message {
		t.Errorf("stderr message = %q, want %q", entry.Message, message)
	}
}

func TestNewLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{Level: "warn", Writer: &buf})
	logger.Info("should-not-appear")
	if buf.Len() > 0 {
		t.Errorf("info log should be filtered at warn level: %s", buf.String())
	}

	logger.Warn("should-appear")
	if buf.Len() == 0 {
		t.Error("warn log should appear at warn level")
	}
}
