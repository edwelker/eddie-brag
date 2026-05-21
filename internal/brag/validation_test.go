package brag

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// MockHTTPClient implements HTTPClient interface for testing
type MockHTTPClient struct {
	statusCode int
	err        error
}

func (m *MockHTTPClient) Head(url string) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

// TestValidateURL tests the ValidateURL function with various HTTP status codes
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		wantValid  bool
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "2xx Success",
			statusCode: 200,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "201 Created",
			statusCode: 201,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "3xx Redirect",
			statusCode: 301,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "308 Permanent Redirect",
			statusCode: 308,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "399 Edge of success range",
			statusCode: 399,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "401 Unauthorized (protected but valid)",
			statusCode: 401,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "403 Forbidden (protected but valid)",
			statusCode: 403,
			err:        nil,
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			err:        nil,
			wantValid:  false,
			wantErr:    false,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			err:        nil,
			wantValid:  false,
			wantErr:    false,
		},
		{
			name:       "502 Bad Gateway",
			statusCode: 502,
			err:        nil,
			wantValid:  false,
			wantErr:    false,
		},
		{
			name:       "503 Service Unavailable",
			statusCode: 503,
			err:        nil,
			wantValid:  false,
			wantErr:    false,
		},
		{
			name:       "Network error (connection timeout)",
			statusCode: 0,
			err:        fmt.Errorf("context deadline exceeded"),
			wantValid:  false,
			wantErr:    true,
			errMsg:     "context deadline exceeded",
		},
		{
			name:       "Network error (connection refused)",
			statusCode: 0,
			err:        fmt.Errorf("connection refused"),
			wantValid:  false,
			wantErr:    true,
			errMsg:     "connection refused",
		},
		{
			name:       "400 Bad Request",
			statusCode: 400,
			err:        nil,
			wantValid:  false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockHTTPClient{
				statusCode: tt.statusCode,
				err:        tt.err,
			}

			got, err := ValidateURL("https://example.com", client)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateURL() error = %v, want to contain %q", err, tt.errMsg)
				}
			}

			if got != tt.wantValid {
				t.Errorf("ValidateURL() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestValidateURL_Redirects tests that redirects (3xx) are handled correctly
func TestValidateURL_Redirects(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantValid  bool
	}{
		{"300 Multiple Choices", 300, true},
		{"301 Moved Permanently", 301, true},
		{"302 Found", 302, true},
		{"303 See Other", 303, true},
		{"304 Not Modified", 304, true},
		{"307 Temporary Redirect", 307, true},
		{"308 Permanent Redirect", 308, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockHTTPClient{statusCode: tt.statusCode}
			got, err := ValidateURL("https://example.com", client)

			if err != nil {
				t.Fatalf("ValidateURL() unexpected error = %v", err)
			}

			if got != tt.wantValid {
				t.Errorf("ValidateURL() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestParseHoursInputValidation tests the ParseHoursInput function
func TestParseHoursInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
		errMsg  string
	}{
		// Plain number inputs
		{
			name:    "Integer hours",
			input:   "2",
			want:    2.0,
			wantErr: false,
		},
		{
			name:    "Decimal hours",
			input:   "1.5",
			want:    1.5,
			wantErr: false,
		},
		{
			name:    "Decimal hours with zero",
			input:   "0.5",
			want:    0.5,
			wantErr: false,
		},
		{
			name:    "Large decimal hours",
			input:   "24.75",
			want:    24.75,
			wantErr: false,
		},
		{
			name:    "Zero hours",
			input:   "0",
			want:    0.0,
			wantErr: false,
		},

		// Hours suffix (h)
		{
			name:    "Integer with h suffix",
			input:   "2h",
			want:    2.0,
			wantErr: false,
		},
		{
			name:    "Decimal with h suffix",
			input:   "1.5h",
			want:    1.5,
			wantErr: false,
		},
		{
			name:    "Single h suffix",
			input:   "1h",
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "Zero with h suffix",
			input:   "0h",
			want:    0.0,
			wantErr: false,
		},

		// Minutes suffix (m)
		{
			name:    "Integer minutes with m suffix",
			input:   "60m",
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "30 minutes to hours",
			input:   "30m",
			want:    0.5,
			wantErr: false,
		},
		{
			name:    "90 minutes to hours",
			input:   "90m",
			want:    1.5,
			wantErr: false,
		},
		{
			name:    "45 minutes to hours",
			input:   "45m",
			want:    0.75,
			wantErr: false,
		},
		{
			name:    "Decimal minutes to hours",
			input:   "30.5m",
			want:    0.5083333333333333,
			wantErr: false,
		},
		{
			name:    "Single minute",
			input:   "1m",
			want:    1.0 / 60.0,
			wantErr: false,
		},

		// Minutes suffix (min)
		{
			name:    "Integer minutes with min suffix",
			input:   "60min",
			want:    1.0,
			wantErr: false,
		},
		{
			name:    "30 minutes with min suffix",
			input:   "30min",
			want:    0.5,
			wantErr: false,
		},
		{
			name:    "Decimal minutes with min suffix",
			input:   "90.5min",
			want:    1.5083333333333334,
			wantErr: false,
		},

		// Whitespace handling
		{
			name:    "Leading whitespace",
			input:   "  2.5",
			want:    2.5,
			wantErr: false,
		},
		{
			name:    "Trailing whitespace",
			input:   "2.5  ",
			want:    2.5,
			wantErr: false,
		},
		{
			name:    "Leading and trailing whitespace",
			input:   "  1.5h  ",
			want:    1.5,
			wantErr: false,
		},
		{
			name:    "Whitespace with minutes",
			input:   "  90m  ",
			want:    1.5,
			wantErr: false,
		},

		// Error cases
		{
			name:    "Invalid number format",
			input:   "abc",
			wantErr: true,
			errMsg:  "invalid number format",
		},
		{
			name:    "Invalid hours format",
			input:   "abch",
			wantErr: true,
			errMsg:  "invalid hours format",
		},
		{
			name:    "Invalid minutes format with m",
			input:   "abcm",
			wantErr: true,
			errMsg:  "invalid minutes format",
		},
		{
			name:    "Invalid minutes format with min",
			input:   "abcmin",
			wantErr: true,
			errMsg:  "invalid minutes format",
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
			errMsg:  "invalid number format",
		},
		{
			name:    "Only whitespace",
			input:   "   ",
			wantErr: true,
			errMsg:  "invalid number format",
		},
		{
			name:    "Invalid suffix",
			input:   "5x",
			wantErr: true,
			errMsg:  "invalid number format",
		},
		{
			name:    "Multiple suffixes",
			input:   "5mh",
			wantErr: true,
			errMsg:  "invalid hours format",
		},
		{
			name:    "Negative number",
			input:   "-5",
			want:    -5.0,
			wantErr: false,
		},
		{
			name:    "Negative with h suffix",
			input:   "-1.5h",
			want:    -1.5,
			wantErr: false,
		},
		{
			name:    "Scientific notation",
			input:   "1e2",
			want:    100.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHoursInput(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHoursInput(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseHoursInput(%q) error = %v, want to contain %q", tt.input, err, tt.errMsg)
				}
			}

			if !tt.wantErr {
				// Use a small epsilon for float comparison
				const epsilon = 1e-9
				if diff := got - tt.want; diff < -epsilon || diff > epsilon {
					t.Errorf("ParseHoursInput(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}

// TestNewDefaultHTTPClient verifies that the default HTTP client is created correctly
func TestNewDefaultHTTPClient(t *testing.T) {
	client := NewDefaultHTTPClient()

	if client == nil {
		t.Fatal("NewDefaultHTTPClient() returned nil")
	}

	if client.client == nil {
		t.Fatal("NewDefaultHTTPClient() created client with nil underlying http.Client")
	}

	if client.client.Timeout != 5*1000*1000*1000 { // 5 seconds in nanoseconds
		t.Errorf("NewDefaultHTTPClient() timeout = %v, want 5s", client.client.Timeout)
	}
}

// BenchmarkParseHoursInput benchmarks the ParseHoursInput function
func BenchmarkParseHoursInput(b *testing.B) {
	inputs := []string{"1.5", "90m", "2h", "45min", "3.25"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_, err := ParseHoursInput(input)
			_ = err
		}
	}
}

// TestDefaultHTTPClient_Head tests the Head method of DefaultHTTPClient
func TestDefaultHTTPClient_Head(t *testing.T) {
	// This test verifies the Head method delegates to the underlying http.Client
	// We'll create a minimal mock to verify the method is callable
	client := NewDefaultHTTPClient()

	if client == nil {
		t.Fatal("NewDefaultHTTPClient() returned nil")
	}

	// Verify the method exists and can be called (with a bad URL for error case)
	_, err := client.Head("http://invalid:url:format")
	if err == nil {
		t.Error("Head() with invalid URL should return an error")
	}
}

// BenchmarkValidateURL benchmarks the ValidateURL function
func BenchmarkValidateURL(b *testing.B) {
	client := &MockHTTPClient{statusCode: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateURL("https://example.com", client)
		_ = err
	}
}
