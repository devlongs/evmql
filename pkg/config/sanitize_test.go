package config

import (
	"testing"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		expected  string
	}{
		{
			name:      "Valid HTTPS URL",
			input:     "https://mainnet.infura.io/v3/key",
			shouldErr: false,
			expected:  "https://mainnet.infura.io/v3/key",
		},
		{
			name:      "Valid HTTP URL",
			input:     "http://localhost:8545",
			shouldErr: false,
			expected:  "http://localhost:8545",
		},
		{
			name:      "Empty URL",
			input:     "",
			shouldErr: false,
			expected:  "",
		},
		{
			name:      "URL with spaces",
			input:     "  https://mainnet.infura.io/v3/key  ",
			shouldErr: false,
			expected:  "https://mainnet.infura.io/v3/key",
		},
		{
			name:      "Invalid scheme",
			input:     "ftp://example.com",
			shouldErr: false,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeURL(tt.input)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSanitizeNetworkName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid name",
			input:    "mainnet",
			expected: "mainnet",
		},
		{
			name:     "Name with dashes",
			input:    "sepolia-testnet",
			expected: "sepolia-testnet",
		},
		{
			name:     "Name with underscores",
			input:    "local_dev",
			expected: "local_dev",
		},
		{
			name:     "Name with invalid characters",
			input:    "test@network!",
			expected: "testnetwork",
		},
		{
			name:     "Name with spaces",
			input:    "test network",
			expected: "testnetwork",
		},
		{
			name:     "Very long name",
			input:    "this_is_a_very_long_network_name_that_exceeds_fifty_characters_limit",
			expected: "this_is_a_very_long_network_name_that_exceeds_fift",
		},
		{
			name:     "Empty after sanitization",
			input:    "@@@###",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeNetworkName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    string
		notContains string
	}{
		{
			name:        "URL with API key",
			input:       "https://mainnet.infura.io/v3/abc123def456ghi789jkl012mno345",
			contains:    "[REDACTED]",
			notContains: "abc123def456ghi789jkl012mno345",
		},
		{
			name:     "No API key",
			input:    "https://localhost:8545",
			contains: "https://localhost:8545",
		},
		{
			name:        "Multiple API keys",
			input:       "key1: abc123def456ghi789jkl012, key2: xyz987wvu654tsr321onm098",
			contains:    "[REDACTED]",
			notContains: "abc123def456ghi789jkl012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactAPIKey(tt.input)

			if tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}

			if tt.notContains != "" && containsString(result, tt.notContains) {
				t.Errorf("Expected result to NOT contain %q, got %q", tt.notContains, result)
			}
		})
	}
}

func TestValidateNodeURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{
			name:      "Valid URL",
			input:     "https://mainnet.infura.io/v3/key",
			shouldErr: false,
		},
		{
			name:      "Empty URL",
			input:     "",
			shouldErr: false,
		},
		{
			name:      "URL with placeholder",
			input:     "https://mainnet.infura.io/v3/YOUR_KEY",
			shouldErr: false,
		},
		{
			name:      "Invalid URL format",
			input:     "not a url at all",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodeURL(tt.input)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
