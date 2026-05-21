package brag

import (
	"testing"
	"time"
)

// Tests already exist in brag_test.go:
// - TestGetWeekRange
// - TestGetMonthRange
// - TestGetYearRange
// - TestGetCurrentPeriod
// - TestParseDuration
// - TestIsUSFederalHoliday
// - TestCountBusinessDays
// - TestResolveDateFlags
// - TestGetTenure

// Additional critical date tests:

func TestGetWeekRange_RoleStartNotMonday(t *testing.T) {
	// Critical: week boundaries based on role start, not calendar weeks
	roleStart := time.Date(2024, 1, 3, 0, 0, 0, 0, time.Local) // Wednesday

	start, end := getWeekRange(roleStart, 1)

	// Week 1 should start on Wednesday 1/3, not Monday 1/1
	if start.Day() != 3 {
		t.Errorf("Week 1 start day = %d, want 3 (role start day)", start.Day())
	}

	if end.Day() != 10 {
		t.Errorf("Week 1 end day = %d, want 10 (7 days after start)", end.Day())
	}
}

func TestGetMonthRange_RoleStartMidMonth(t *testing.T) {
	// Critical: month boundaries based on role start day, not calendar months
	roleStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local) // Mid-January

	start, end := getMonthRange(roleStart, 1)

	// Month 1 should be Jan 15 - Feb 15, not Jan 1 - Jan 31
	if start.Day() != 15 {
		t.Errorf("Month 1 start day = %d, want 15 (role start day)", start.Day())
	}

	if end.Month() != time.February || end.Day() != 15 {
		t.Errorf("Month 1 end = %v, want Feb 15", end)
	}
}

func TestCountBusinessDays_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "Same day",
			start:    time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local), // Monday
			end:      time.Date(2026, 5, 18, 0, 0, 0, 0, time.Local),
			expected: 1,
		},
		{
			name:     "Weekend only",
			start:    time.Date(2026, 5, 23, 0, 0, 0, 0, time.Local), // Saturday
			end:      time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), // Sunday
			expected: 0,
		},
		{
			name:     "Holiday only (Memorial Day 2026)",
			start:    time.Date(2026, 5, 25, 0, 0, 0, 0, time.Local), // Memorial Day Monday
			end:      time.Date(2026, 5, 25, 0, 0, 0, 0, time.Local),
			expected: 0,
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

func TestIsUSFederalHoliday_ObservedDates(t *testing.T) {
	// When a fixed holiday falls on weekend, it's often observed on nearest weekday
	// But our function checks the actual date, not observed date
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "July 4th 2026 (Saturday)",
			date:     time.Date(2026, 7, 4, 0, 0, 0, 0, time.Local),
			expected: true, // Still a holiday even though weekend
		},
		{
			name:     "Christmas 2026 (Friday)",
			date:     time.Date(2026, 12, 25, 0, 0, 0, 0, time.Local),
			expected: true,
		},
		{
			name:     "Day after Christmas 2026 (Saturday)",
			date:     time.Date(2026, 12, 26, 0, 0, 0, 0, time.Local),
			expected: false, // Not the actual holiday
		},
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

func TestGetTenure_RecentStart(t *testing.T) {
	// Should not panic or return negative values for very recent start dates
	roleStart := time.Now().AddDate(0, 0, -1) // Started yesterday
	tenure := getTenure(roleStart)

	if tenure == "" {
		t.Error("getTenure() returned empty string")
	}

	// Should mention "since" and include the date
	if len(tenure) < 10 {
		t.Errorf("getTenure() = %q, seems too short", tenure)
	}
}

func TestParseDuration_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantDays  int
		wantHours int64
		wantErr   bool
	}{
		// Valid single-unit formats
		{"1 day", "1d", 1, 24, false},
		{"0 days edge case", "0d", 0, 0, false},
		{"large days", "365d", 365, 365 * 24, false},
		{"1 week", "1w", 7, 7 * 24, false},
		{"multiple weeks", "4w", 28, 28 * 24, false},
		{"1 month", "1m", 30, 30 * 24, false},
		{"multiple months", "12m", 360, 360 * 24, false},
		{"1 year", "1y", 365, 365 * 24, false},
		{"multiple years", "3y", 1095, 1095 * 24, false},

		// Invalid formats - no match
		{"no unit", "5", 0, 0, true},
		{"invalid unit", "5x", 0, 0, true},
		{"empty string", "", 0, 0, true},
		{"letters only", "abc", 0, 0, true},
		{"negative sign", "-5d", 0, 0, true},
		{"decimal", "1.5d", 0, 0, true},
		{"space in input", "5 d", 0, 0, true},
		{"unit only", "d", 0, 0, true},
		{"multiple units", "1d2w", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotDays := int(got.Hours() / 24)
				if gotDays != tt.wantDays {
					t.Errorf("parseDuration(%q) days = %d, want %d", tt.input, gotDays, tt.wantDays)
				}
				gotHours := got.Hours()
				if int64(gotHours) != tt.wantHours {
					t.Errorf("parseDuration(%q) hours = %d, want %d", tt.input, int64(gotHours), tt.wantHours)
				}
			}
		})
	}
}
