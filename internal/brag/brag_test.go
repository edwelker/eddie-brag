package brag

import (
	"net/http"
	"testing"
	"time"
)

// Mock HTTP client for testing
type mockHTTPClient struct {
	statusCode int
	err        error
}

func (m *mockHTTPClient) Head(url string) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       http.NoBody,
	}, nil
}

func TestValidateURL_Success(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantValid  bool
	}{
		{"200 OK", 200, true},
		{"302 Redirect", 302, true},
		{"401 Unauthorized", 401, true},
		{"403 Forbidden", 403, true},
		{"404 Not Found", 404, false},
		{"500 Internal Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockHTTPClient{statusCode: tt.statusCode}
			valid, err := ValidateURL("http://example.com", client)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if valid != tt.wantValid {
				t.Errorf("ValidateURL() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestParseHoursInput(t *testing.T) {
	tests := []struct {
		input     string
		wantHours float64
		wantErr   bool
	}{
		{"1.5", 1.5, false},
		{"90m", 1.5, false},
		{"90min", 1.5, false},
		{"2h", 2.0, false},
		{"0", 0.0, false},
		{"120m", 2.0, false},
		{"invalid", 0.0, true},
		{"", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseHoursInput(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHoursInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantHours {
				t.Errorf("ParseHoursInput(%q) = %v, want %v", tt.input, got, tt.wantHours)
			}
		})
	}
}

func TestGetWeekRange(t *testing.T) {
	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		weekNum int
		wantStart string
		wantEnd string
	}{
		{1, "2024-01-01", "2024-01-08"},
		{2, "2024-01-08", "2024-01-15"},
		{4, "2024-01-22", "2024-01-29"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.weekNum)), func(t *testing.T) {
			start, end := getWeekRange(roleStart, tt.weekNum)

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("getWeekRange(%d) start = %s, want %s", tt.weekNum, gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("getWeekRange(%d) end = %s, want %s", tt.weekNum, gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestGetMonthRange(t *testing.T) {
	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		monthNum int
		wantStart string
		wantEnd string
	}{
		{1, "2024-01-01", "2024-02-01"},
		{2, "2024-02-01", "2024-03-01"},
		{12, "2024-12-01", "2025-01-01"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.monthNum)), func(t *testing.T) {
			start, end := getMonthRange(roleStart, tt.monthNum)

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("getMonthRange(%d) start = %s, want %s", tt.monthNum, gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("getMonthRange(%d) end = %s, want %s", tt.monthNum, gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		wantDays int
		wantErr bool
	}{
		{"7d", 7, false},
		{"2w", 14, false},
		{"1m", 30, false},
		{"30d", 30, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotDays := int(got.Hours() / 24)
				if gotDays != tt.wantDays {
					t.Errorf("parseDuration(%q) = %d days, want %d days", tt.input, gotDays, tt.wantDays)
				}
			}
		})
	}
}

func TestFilterAfter(t *testing.T) {
	entries := []Entry{
		{ID: 1, StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
		{ID: 2, StartDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local)},
		{ID: 3, StartDate: time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local)},
	}

	cutoff := time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local)
	filtered := filterAfter(entries, cutoff)

	if len(filtered) != 2 {
		t.Errorf("filterAfter() returned %d entries, want 2", len(filtered))
	}

	if filtered[0].ID != 2 || filtered[1].ID != 3 {
		t.Errorf("filterAfter() returned wrong entries")
	}
}

func TestGroupByBucket(t *testing.T) {
	entries := []Entry{
		{ID: 1, Bucket: "Delivery"},
		{ID: 2, Bucket: "Process"},
		{ID: 3, Bucket: "Delivery"},
	}

	grouped := groupByBucket(entries)

	if len(grouped) != 2 {
		t.Errorf("groupByBucket() returned %d buckets, want 2", len(grouped))
	}

	if len(grouped["Delivery"]) != 2 {
		t.Errorf("Delivery bucket has %d entries, want 2", len(grouped["Delivery"]))
	}

	if len(grouped["Process"]) != 1 {
		t.Errorf("Process bucket has %d entries, want 1", len(grouped["Process"]))
	}
}
