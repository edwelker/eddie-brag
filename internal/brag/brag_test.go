package brag

import (
	"fmt"
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

func TestValidateURL_ErrorPath(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "network timeout",
			err:       fmt.Errorf("context deadline exceeded"),
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "invalid URL",
			err:       fmt.Errorf("unsupported protocol scheme"),
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "DNS failure",
			err:       fmt.Errorf("no such host"),
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockHTTPClient{err: tt.err}
			valid, err := ValidateURL("http://example.com", client)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
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
		{"0m", 0.0, false},
		{"120m", 2.0, false},
		{"invalid", 0.0, true},
		{"", 0.0, true},
		{"-5", -5.0, false}, // Negative is parsed but should be validated elsewhere
		{"9999999", 9999999.0, false},
		{"abc123", 0.0, true},
		{"-10m", -10.0 / 60.0, false}, // Negative minutes parsed
		{"invalidh", 0.0, true},       // Invalid hours format
		{"badmin", 0.0, true},         // Invalid minutes format
		{"xyzm", 0.0, true},           // Invalid minutes format
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
		weekNum   int
		wantStart string
		wantEnd   string
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
		monthNum  int
		wantStart string
		wantEnd   string
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
		input    string
		wantDays int
		wantErr  bool
	}{
		{"7d", 7, false},
		{"2w", 14, false},
		{"1m", 30, false},
		{"30d", 30, false},
		{"1y", 365, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"10x", 0, true}, // Invalid suffix
		{"-5d", 0, true}, // Negative number
		{"5", 0, true},   // Missing unit
		{"abc", 0, true}, // Non-numeric
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

func TestIsUSFederalHoliday(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		// Fixed holidays
		{"New Year's Day 2026", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), true},
		{"July 4th 2026", time.Date(2026, 7, 4, 0, 0, 0, 0, time.Local), true},
		{"Veterans Day 2026", time.Date(2026, 11, 11, 0, 0, 0, 0, time.Local), true},
		{"Christmas 2026", time.Date(2026, 12, 25, 0, 0, 0, 0, time.Local), true},
		{"Juneteenth 2026", time.Date(2026, 6, 19, 0, 0, 0, 0, time.Local), true},

		// Floating holidays - 2026
		{"MLK Day 2026 (Jan 19)", time.Date(2026, 1, 19, 0, 0, 0, 0, time.Local), true},
		{"Presidents Day 2026 (Feb 16)", time.Date(2026, 2, 16, 0, 0, 0, 0, time.Local), true},
		{"Memorial Day 2026 (May 25)", time.Date(2026, 5, 25, 0, 0, 0, 0, time.Local), true},
		{"Labor Day 2026 (Sep 7)", time.Date(2026, 9, 7, 0, 0, 0, 0, time.Local), true},
		{"Thanksgiving 2026 (Nov 26)", time.Date(2026, 11, 26, 0, 0, 0, 0, time.Local), true},

		// Not holidays
		{"Random Tuesday", time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local), false},
		{"Day before MLK", time.Date(2026, 1, 18, 0, 0, 0, 0, time.Local), false},
		{"Day after Thanksgiving", time.Date(2026, 11, 27, 0, 0, 0, 0, time.Local), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUSFederalHoliday(tt.date)
			if got != tt.expected {
				t.Errorf("isUSFederalHoliday(%s) = %v, want %v",
					tt.date.Format("2006-01-02"), got, tt.expected)
			}
		})
	}
}

func TestCountBusinessDays(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "Single weekday",
			start:    time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local), // Monday
			end:      time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local),
			expected: 1,
		},
		{
			name:     "Monday through Friday (5 days)",
			start:    time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local), // Monday
			end:      time.Date(2026, 5, 22, 0, 0, 0, 0, time.Local), // Friday
			expected: 5,
		},
		{
			name:     "Week including weekend (Mon-Sun = 5 business days)",
			start:    time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local), // Monday
			end:      time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), // Sunday
			expected: 5,
		},
		{
			name:     "Week with holiday (Memorial Day May 25)",
			start:    time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local), // Monday
			end:      time.Date(2026, 5, 29, 0, 0, 0, 0, time.Local), // Friday
			expected: 9,                                              // Mon 18-Fri 22 (5) + Tue 26-Fri 29 (4), Memorial Day Mon 25 excluded
		},
		{
			name:     "March 30 to May 20 2026",
			start:    time.Date(2026, 3, 30, 0, 0, 0, 0, time.Local),
			end:      time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local),
			expected: 38, // Real calculation for your role start
		},
		{
			name:     "Span including July 4th weekend",
			start:    time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local),  // Wednesday
			end:      time.Date(2026, 7, 10, 0, 0, 0, 0, time.Local), // Friday
			expected: 8,                                              // Wed 1, Thu 2, Fri 3 (3) + Mon 7, Tue 8, Wed 9, Thu 10, Fri 10 (5) = 8, July 4 is Sat (already excluded)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countBusinessDays(tt.start, tt.end)
			if got != tt.expected {
				t.Errorf("countBusinessDays(%s to %s) = %d, want %d",
					tt.start.Format("2006-01-02"),
					tt.end.Format("2006-01-02"),
					got, tt.expected)
			}
		})
	}
}

func TestGetYearRange(t *testing.T) {
	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		yearNum   int
		wantStart string
		wantEnd   string
	}{
		{1, "2024-01-01", "2025-01-01"},
		{2, "2025-01-01", "2026-01-01"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.yearNum)), func(t *testing.T) {
			start, end := getYearRange(roleStart, tt.yearNum)

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("getYearRange(%d) start = %s, want %s", tt.yearNum, gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("getYearRange(%d) end = %s, want %s", tt.yearNum, gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestGetCurrentPeriod(t *testing.T) {
	// Test with a known role start date
	roleStart := time.Date(2026, 3, 30, 0, 0, 0, 0, time.Local)

	// As of 2026-05-20, we're 51 days in
	// Week: 51/7 + 1 = 8
	// Month: 51/30 + 1 = 2
	// Year: 51/365 + 1 = 1

	week := GetCurrentWeek(roleStart)
	if week < 7 || week > 9 {
		t.Errorf("GetCurrentWeek() = %d, expected between 7-9", week)
	}

	month := GetCurrentMonth(roleStart)
	if month < 1 || month > 3 {
		t.Errorf("GetCurrentMonth() = %d, expected between 1-3", month)
	}

	year := GetCurrentYear(roleStart)
	if year != 1 {
		t.Errorf("GetCurrentYear() = %d, want 1", year)
	}
}

func TestFilterByDateRange(t *testing.T) {
	entries := []Entry{
		{ID: 1, StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 3, 0, 0, 0, 0, time.Local)},
		{ID: 2, StartDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 12, 0, 0, 0, 0, time.Local)},
		{ID: 3, StartDate: time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 22, 0, 0, 0, 0, time.Local)},
	}

	start := time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local)
	end := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)
	filtered := filterByDateRange(entries, start, end)

	if len(filtered) != 1 {
		t.Errorf("filterByDateRange() returned %d entries, want 1", len(filtered))
	}

	if len(filtered) > 0 && filtered[0].ID != 2 {
		t.Errorf("filterByDateRange() returned entry #%d, want #2", filtered[0].ID)
	}
}

func TestResolveDateFlags(t *testing.T) {
	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name          string
		weekNum       int
		monthNum      int
		startDateStr  string
		endDateStr    string
		wantStartDate string
		wantEndDate   string
		wantErr       bool
	}{
		{
			name:          "Week number",
			weekNum:       1,
			wantStartDate: "2024-01-01",
			wantEndDate:   "2024-01-08",
		},
		{
			name:          "Month number",
			monthNum:      2,
			wantStartDate: "2024-02-01",
			wantEndDate:   "2024-03-01",
		},
		{
			name:          "Explicit dates",
			startDateStr:  "2024-01-15",
			endDateStr:    "2024-01-20",
			wantStartDate: "2024-01-15",
			wantEndDate:   "2024-01-20",
		},
		{
			name:          "No flags defaults to today",
			wantStartDate: time.Now().Format("2006-01-02"),
			wantEndDate:   time.Now().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ResolveDateFlags(roleStart, tt.weekNum, tt.monthNum, tt.startDateStr, tt.endDateStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveDateFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotStart := start.Format("2006-01-02")
				gotEnd := end.Format("2006-01-02")

				if gotStart != tt.wantStartDate {
					t.Errorf("ResolveDateFlags() start = %s, want %s", gotStart, tt.wantStartDate)
				}

				if gotEnd != tt.wantEndDate {
					t.Errorf("ResolveDateFlags() end = %s, want %s", gotEnd, tt.wantEndDate)
				}
			}
		})
	}
}

func TestEntry_CalculateCompleteness(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
		want  int
	}{
		{
			name: "Empty entry",
			entry: Entry{
				Description: "",
				Evidence:    "",
			},
			want: 0,
		},
		{
			name: "Base fields only (description + evidence)",
			entry: Entry{
				Description: "Did something",
				Evidence:    "http://example.com",
			},
			want: 40,
		},
		{
			name: "Base + hours saved",
			entry: Entry{
				Description: "Did something",
				Evidence:    "http://example.com",
				HoursSaved:  ptrFloat64(10.0),
			},
			want: 55,
		},
		{
			name: "Base + business metric",
			entry: Entry{
				Description:    "Did something",
				Evidence:       "http://example.com",
				BusinessMetric: "Saved 10 hours",
			},
			want: 55,
		},
		{
			name: "Base + strategic align",
			entry: Entry{
				Description:    "Did something",
				Evidence:       "http://example.com",
				StrategicAlign: "Developer velocity",
			},
			want: 55,
		},
		{
			name: "Base + peer recognition",
			entry: Entry{
				Description:     "Did something",
				Evidence:        "http://example.com",
				PeerRecognition: "Team praised it",
			},
			want: 55,
		},
		{
			name: "All fields complete",
			entry: Entry{
				Description:     "Did something",
				Evidence:        "http://example.com",
				HoursSaved:      ptrFloat64(10.0),
				BusinessMetric:  "Saved 10 hours",
				StrategicAlign:  "Developer velocity",
				PeerRecognition: "Team praised it",
			},
			want: 100,
		},
		{
			name: "Missing evidence marked as [missing]",
			entry: Entry{
				Description: "Did something",
				Evidence:    "[missing]",
			},
			want: 20, // Only description counts
		},
		{
			name: "All enrichment fields but no base",
			entry: Entry{
				HoursSaved:      ptrFloat64(10.0),
				BusinessMetric:  "Saved 10 hours",
				StrategicAlign:  "Developer velocity",
				PeerRecognition: "Team praised it",
			},
			want: 60, // 15 * 4 enrichment fields
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.CalculateCompleteness()
			if got != tt.want {
				t.Errorf("CalculateCompleteness() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetTenure(t *testing.T) {
	tests := []struct {
		name      string
		roleStart time.Time
		contains  []string // Substrings that should appear in output
	}{
		{
			name:      "Recent start (1 week)",
			roleStart: time.Now().AddDate(0, 0, -7),
			contains:  []string{"Week", "since"},
		},
		{
			name:      "Several months ago",
			roleStart: time.Now().AddDate(0, -3, 0),
			contains:  []string{"Week", "Month", "since"},
		},
		{
			name:      "Over a year ago",
			roleStart: time.Now().AddDate(-1, -1, 0),
			contains:  []string{"Week", "Month", "years", "since"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTenure(tt.roleStart)
			for _, substr := range tt.contains {
				if !contains(got, substr) {
					t.Errorf("getTenure() = %q, missing substring %q", got, substr)
				}
			}
		})
	}
}

// Helper functions
func ptrFloat64(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestResolveDateFlags_ExplicitDates(t *testing.T) {
	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name      string
		startStr  string
		endStr    string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{
			name:      "Both dates provided",
			startStr:  "2024-02-01",
			endStr:    "2024-02-15",
			wantStart: "2024-02-01",
			wantEnd:   "2024-02-15",
			wantErr:   false,
		},
		{
			name:      "Only start date",
			startStr:  "2024-03-15",
			endStr:    "",
			wantStart: "2024-03-15",
			wantEnd:   time.Now().Format("2006-01-02"),
			wantErr:   false,
		},
		{
			name:      "Invalid start date",
			startStr:  "not-a-date",
			endStr:    "",
			wantStart: "",
			wantEnd:   "",
			wantErr:   true,
		},
		{
			name:      "Invalid end date",
			startStr:  "2024-01-01",
			endStr:    "invalid",
			wantStart: "",
			wantEnd:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ResolveDateFlags(roleStart, 0, 0, tt.startStr, tt.endStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveDateFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotStart := start.Format("2006-01-02")
				gotEnd := end.Format("2006-01-02")

				if gotStart != tt.wantStart {
					t.Errorf("ResolveDateFlags() start = %s, want %s", gotStart, tt.wantStart)
				}

				if gotEnd != tt.wantEnd {
					t.Errorf("ResolveDateFlags() end = %s, want %s", gotEnd, tt.wantEnd)
				}
			}
		})
	}
}
