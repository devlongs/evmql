package logger

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger returned nil")
	}
}

func TestSetLogger(t *testing.T) {
	var buf bytes.Buffer

	customLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	SetLogger(customLogger)

	retrievedLogger := GetLogger()
	if retrievedLogger != customLogger {
		t.Error("SetLogger did not set the custom logger")
	}
}

func TestDebug(t *testing.T) {
	var buf bytes.Buffer

	testLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	SetLogger(testLogger)

	Debug("test debug message", "key", "value")

	if buf.Len() == 0 {
		t.Error("Debug message was not logged")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("test debug message")) {
		t.Error("Debug message not found in output")
	}
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer

	testLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	SetLogger(testLogger)

	Info("test info message", "key", "value")

	if buf.Len() == 0 {
		t.Error("Info message was not logged")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("test info message")) {
		t.Error("Info message not found in output")
	}
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer

	testLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	SetLogger(testLogger)

	Warn("test warn message", "key", "value")

	if buf.Len() == 0 {
		t.Error("Warn message was not logged")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("test warn message")) {
		t.Error("Warn message not found in output")
	}
}

func TestError(t *testing.T) {
	var buf bytes.Buffer

	testLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	SetLogger(testLogger)

	Error("test error message", "key", "value")

	if buf.Len() == 0 {
		t.Error("Error message was not logged")
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("test error message")) {
		t.Error("Error message not found in output")
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  slog.Level
		logFunc   func(string, ...any)
		message   string
		shouldLog bool
	}{
		{
			name:      "Debug at Debug level",
			logLevel:  slog.LevelDebug,
			logFunc:   Debug,
			message:   "debug message",
			shouldLog: true,
		},
		{
			name:      "Debug at Info level",
			logLevel:  slog.LevelInfo,
			logFunc:   Debug,
			message:   "debug message",
			shouldLog: false,
		},
		{
			name:      "Info at Info level",
			logLevel:  slog.LevelInfo,
			logFunc:   Info,
			message:   "info message",
			shouldLog: true,
		},
		{
			name:      "Info at Warn level",
			logLevel:  slog.LevelWarn,
			logFunc:   Info,
			message:   "info message",
			shouldLog: false,
		},
		{
			name:      "Warn at Warn level",
			logLevel:  slog.LevelWarn,
			logFunc:   Warn,
			message:   "warn message",
			shouldLog: true,
		},
		{
			name:      "Error at Error level",
			logLevel:  slog.LevelError,
			logFunc:   Error,
			message:   "error message",
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			testLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: tt.logLevel,
			}))

			SetLogger(testLogger)
			tt.logFunc(tt.message, "test", "value")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, got hasOutput=%v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestLogWithMultipleFields(t *testing.T) {
	var buf bytes.Buffer

	testLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	SetLogger(testLogger)

	Info("test message",
		"field1", "value1",
		"field2", 123,
		"field3", true)

	output := buf.String()

	expectedSubstrings := []string{
		"test message",
		"field1",
		"value1",
		"field2",
		"123",
		"field3",
		"true",
	}

	for _, substr := range expectedSubstrings {
		if !bytes.Contains([]byte(output), []byte(substr)) {
			t.Errorf("Expected substring '%s' not found in output: %s", substr, output)
		}
	}
}
