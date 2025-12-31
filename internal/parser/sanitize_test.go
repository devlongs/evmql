package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal query",
			input:    "SELECT BALANCE FROM 0x123",
			expected: "SELECT BALANCE FROM 0x123",
		},
		{
			name:     "Extra whitespace",
			input:    "SELECT   BALANCE    FROM   0x123",
			expected: "SELECT BALANCE FROM 0x123",
		},
		{
			name:     "Newlines and tabs",
			input:    "SELECT\nBALANCE\tFROM\n0x123",
			expected: "SELECT BALANCE FROM 0x123",
		},
		{
			name:     "Control characters",
			input:    "SELECT\x00BALANCE\x01FROM\x020x123",
			expected: "SELECT BALANCE FROM 0x123",
		},
		{
			name:     "Leading and trailing spaces",
			input:    "  SELECT BALANCE FROM 0x123  ",
			expected: "SELECT BALANCE FROM 0x123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateNoSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid query",
			input:    "SELECT BALANCE FROM 0x123",
			expected: true,
		},
		{
			name:     "SQL injection attempt - UNION SELECT",
			input:    "SELECT BALANCE FROM 0x123 UNION SELECT *",
			expected: false,
		},
		{
			name:     "SQL injection attempt - DROP",
			input:    "; DROP TABLE users",
			expected: false,
		},
		{
			name:     "Script injection",
			input:    "<script>alert('xss')</script>",
			expected: false,
		},
		{
			name:     "JavaScript injection",
			input:    "javascript:alert(1)",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateNoSQLInjection(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input %q", tt.expected, result, tt.input)
			}
		})
	}
}

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase with 0x",
			input:    "0xabcdef1234567890abcdef1234567890abcdef12",
			expected: "0xabcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:     "Uppercase with 0x",
			input:    "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			expected: "0xabcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:     "Without 0x prefix",
			input:    "abcdef1234567890abcdef1234567890abcdef12",
			expected: "0xabcdef1234567890abcdef1234567890abcdef12",
		},
		{
			name:     "With spaces",
			input:    "  0xABCDEF1234567890ABCDEF1234567890ABCDEF12  ",
			expected: "0xabcdef1234567890abcdef1234567890abcdef12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAddress(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateAddressFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid address",
			input:    "0xabcdef1234567890abcdef1234567890abcdef12",
			expected: true,
		},
		{
			name:     "Too short",
			input:    "0xabcdef",
			expected: false,
		},
		{
			name:     "Too long",
			input:    "0xabcdef1234567890abcdef1234567890abcdef12345",
			expected: false,
		},
		{
			name:     "Missing 0x prefix",
			input:    "abcdef1234567890abcdef1234567890abcdef12",
			expected: false,
		},
		{
			name:     "Invalid characters",
			input:    "0xabcdefg234567890abcdef1234567890abcdef12",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAddressFormat(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for input %q", tt.expected, result, tt.input)
			}
		})
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       error
		contains    []string
		notContains []string
	}{
		{
			name:        "Nil error",
			input:       nil,
			contains:    []string{""},
			notContains: []string{},
		},
		{
			name:        "Error with YOUR_KEY",
			input:       errors.New("node URL contains YOUR_KEY placeholder"),
			contains:    []string{"[REDACTED]"},
			notContains: []string{"YOUR_KEY"},
		},
		{
			name:        "Error with API key pattern",
			input:       errors.New("failed with key abc123def456ghi789jkl012mno345pqr678"),
			contains:    []string{"[API_KEY_REDACTED]"},
			notContains: []string{"abc123def456ghi789jkl012mno345pqr678"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeErrorMessage(tt.input)

			for _, str := range tt.contains {
				if !strings.Contains(result, str) {
					t.Errorf("Expected result to contain %q, got %q", str, result)
				}
			}

			for _, str := range tt.notContains {
				if strings.Contains(result, str) {
					t.Errorf("Expected result to NOT contain %q, got %q", str, result)
				}
			}
		})
	}
}

func TestTruncateForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "Short string",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "Exact length",
			input:    "exactlen",
			maxLen:   8,
			expected: "exactlen",
		},
		{
			name:     "Too long",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is a ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateForDisplay(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
