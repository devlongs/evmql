package config

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	apiKeyPattern = regexp.MustCompile(`[a-zA-Z0-9]{20,}`)
	urlPattern    = regexp.MustCompile(`^https?://`)
)

// SanitizeURL validates and sanitizes a URL
func SanitizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return "", nil
	}

	if !urlPattern.MatchString(rawURL) {
		return "", nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", nil
	}

	if parsedURL.Host == "" {
		return "", nil
	}

	return parsedURL.String(), nil
}

// SanitizeNetworkName removes dangerous characters from network names
func SanitizeNetworkName(name string) string {
	name = strings.TrimSpace(name)

	var cleaned strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			cleaned.WriteRune(r)
		}
	}

	result := cleaned.String()
	if len(result) > 50 {
		result = result[:50]
	}

	return result
}

// RedactAPIKey replaces API keys in strings with a redacted placeholder
func RedactAPIKey(s string) string {
	return apiKeyPattern.ReplaceAllString(s, "[REDACTED]")
}

// ValidateNodeURL checks if a node URL is properly formatted and doesn't contain placeholders
func ValidateNodeURL(nodeURL string) error {
	if nodeURL == "" {
		return nil
	}

	if strings.Contains(nodeURL, "YOUR_KEY") || strings.Contains(nodeURL, "YOUR_API_KEY") {
		return nil
	}

	parsedURL, err := url.Parse(nodeURL)
	if err != nil {
		return err
	}

	// Validate it has a scheme and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return url.InvalidHostError(nodeURL)
	}

	return nil
}
