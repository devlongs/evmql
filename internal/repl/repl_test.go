package repl

import (
	"testing"
)

func TestConfig_Defaults(t *testing.T) {
	config := Config{
		HistoryFile:   "/tmp/history",
		MaxHistoryLen: 1000,
		ColorOutput:   true,
		ShowTimings:   true,
	}

	if config.HistoryFile != "/tmp/history" {
		t.Errorf("Expected history file /tmp/history, got %s", config.HistoryFile)
	}

	if config.MaxHistoryLen != 1000 {
		t.Errorf("Expected max history len 1000, got %d", config.MaxHistoryLen)
	}

	if !config.ColorOutput {
		t.Error("Expected color output to be true")
	}

	if !config.ShowTimings {
		t.Error("Expected show timings to be true")
	}
}

func TestConfig_DisabledFeatures(t *testing.T) {
	config := Config{
		HistoryFile:   "",
		MaxHistoryLen: 0,
		ColorOutput:   false,
		ShowTimings:   false,
	}

	if config.HistoryFile != "" {
		t.Errorf("Expected empty history file, got %s", config.HistoryFile)
	}

	if config.MaxHistoryLen != 0 {
		t.Errorf("Expected max history len 0, got %d", config.MaxHistoryLen)
	}

	if config.ColorOutput {
		t.Error("Expected color output to be false")
	}

	if config.ShowTimings {
		t.Error("Expected show timings to be false")
	}
}

func TestConfig_CustomValues(t *testing.T) {
	tests := []struct {
		name          string
		historyFile   string
		maxHistoryLen int
		colorOutput   bool
		showTimings   bool
	}{
		{
			name:          "Full featured",
			historyFile:   "/home/user/.evmql_history",
			maxHistoryLen: 5000,
			colorOutput:   true,
			showTimings:   true,
		},
		{
			name:          "Minimal",
			historyFile:   "",
			maxHistoryLen: 100,
			colorOutput:   false,
			showTimings:   false,
		},
		{
			name:          "Mixed",
			historyFile:   "/tmp/test_history",
			maxHistoryLen: 500,
			colorOutput:   true,
			showTimings:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				HistoryFile:   tt.historyFile,
				MaxHistoryLen: tt.maxHistoryLen,
				ColorOutput:   tt.colorOutput,
				ShowTimings:   tt.showTimings,
			}

			if config.HistoryFile != tt.historyFile {
				t.Errorf("Expected history file %s, got %s", tt.historyFile, config.HistoryFile)
			}

			if config.MaxHistoryLen != tt.maxHistoryLen {
				t.Errorf("Expected max history len %d, got %d", tt.maxHistoryLen, config.MaxHistoryLen)
			}

			if config.ColorOutput != tt.colorOutput {
				t.Errorf("Expected color output %v, got %v", tt.colorOutput, config.ColorOutput)
			}

			if config.ShowTimings != tt.showTimings {
				t.Errorf("Expected show timings %v, got %v", tt.showTimings, config.ShowTimings)
			}
		})
	}
}
