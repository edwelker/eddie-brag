package brag

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPClient interface for URL validation (allows mocking in tests)
type HTTPClient interface {
	Head(url string) (*http.Response, error)
}

// DefaultHTTPClient wraps http.Client
type DefaultHTTPClient struct {
	client *http.Client
}

func (c *DefaultHTTPClient) Head(url string) (*http.Response, error) {
	return c.client.Head(url)
}

// NewDefaultHTTPClient creates a default HTTP client with timeout
func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ValidateURL checks if a URL is reachable
func ValidateURL(url string, client HTTPClient) (bool, error) {
	resp, err := client.Head(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 200-399: valid
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	}

	// 401, 403: protected but valid
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return true, nil
	}

	// 404, 500+: warn
	return false, nil
}

// ParseHoursInput parses hours from string (handles "1.5", "90m", "2h")
func ParseHoursInput(input string) (float64, error) {
	input = strings.TrimSpace(input)

	// Check for minutes
	if strings.HasSuffix(input, "m") || strings.HasSuffix(input, "min") {
		numStr := strings.TrimSuffix(strings.TrimSuffix(input, "min"), "m")
		mins, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes format")
		}
		return mins / 60.0, nil
	}

	// Check for hours
	if strings.HasSuffix(input, "h") {
		numStr := strings.TrimSuffix(input, "h")
		hours, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid hours format")
		}
		return hours, nil
	}

	// Plain number (hours)
	hours, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format")
	}

	return hours, nil
}
