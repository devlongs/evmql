package parser

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	addressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	// Pattern for dangerous SQL/script injections (excluding valid EVMQL SELECT)
	dangerousPatterns = regexp.MustCompile(`(?i)(union\s+select|;\s*drop|;\s*insert|;\s*update|;\s*delete|;\s*create|;\s*alter|exec\s*\(|<script|javascript:|eval\s*\()`)
)

// SanitizeInput removes potentially dangerous characters and normalizes whitespace
func SanitizeInput(input string) string {
	var cleaned strings.Builder
	cleaned.Grow(len(input))

	prevWasControl := false
	for _, r := range input {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			// Mark that we hit a control char to add space if next char is printable
			prevWasControl = true
			continue
		} else if unicode.IsSpace(r) {
			cleaned.WriteRune(' ')
			prevWasControl = false
		} else if unicode.IsPrint(r) {
			// Add space before this char if previous was a control char
			if prevWasControl {
				cleaned.WriteRune(' ')
			}
			cleaned.WriteRune(r)
			prevWasControl = false
		}
	}

	result := strings.TrimSpace(cleaned.String())
	result = normalizeWhitespace(result)

	return result
}

// normalizeWhitespace replaces multiple consecutive spaces with a single space
func normalizeWhitespace(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	lastWasSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastWasSpace {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteRune(r)
			lastWasSpace = false
		}
	}

	return result.String()
}

// ValidateNoSQLInjection checks for dangerous SQL/script injection patterns
// Allows valid EVMQL syntax like "SELECT BALANCE FROM address"
func ValidateNoSQLInjection(input string) bool {
	return !dangerousPatterns.MatchString(input)
}

// NormalizeAddress ensures address is lowercase with 0x prefix
func NormalizeAddress(addr string) string {
	addr = strings.TrimSpace(addr)
	addr = strings.ToLower(addr)

	if !strings.HasPrefix(addr, "0x") {
		addr = "0x" + addr
	}

	return addr
}

// ValidateAddressFormat checks if address matches expected format
func ValidateAddressFormat(addr string) bool {
	return addressPattern.MatchString(addr)
}

// SanitizeErrorMessage removes sensitive information from error messages
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	msg = strings.ReplaceAll(msg, "YOUR_KEY", "[REDACTED]")
	msg = strings.ReplaceAll(msg, "YOUR_API_KEY", "[REDACTED]")

	apiKeyPattern := regexp.MustCompile(`[a-zA-Z0-9]{32,}`)
	msg = apiKeyPattern.ReplaceAllString(msg, "[API_KEY_REDACTED]")

	return msg
}

// TruncateForDisplay truncates long strings for safe display in error messages
func TruncateForDisplay(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
