package brag

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ResolveDateFlags resolves date flags into start/end times
func ResolveDateFlags(roleStartDate time.Time, weekNum, monthNum int, startStr, endStr string) (time.Time, time.Time, error) {
	now := time.Now().In(time.Local)
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Default: today
	start := nowDate
	end := nowDate

	if weekNum > 0 {
		start, end = getWeekRange(roleStartDate, weekNum)
	} else if monthNum > 0 {
		start, end = getMonthRange(roleStartDate, monthNum)
	} else {
		// Parse explicit dates if provided
		if startStr != "" {
			var err error
			start, err = time.ParseInLocation("2006-01-02", startStr, time.Local)
			if err != nil {
				return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
			}
		}
		if endStr != "" {
			var err error
			end, err = time.ParseInLocation("2006-01-02", endStr, time.Local)
			if err != nil {
				return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
			}
		}
	}

	return start, end, nil
}

func getWeekRange(roleStart time.Time, weekNum int) (time.Time, time.Time) {
	start := roleStart.AddDate(0, 0, (weekNum-1)*7)
	end := start.AddDate(0, 0, 7)
	return start, end
}

func getMonthRange(roleStart time.Time, monthNum int) (time.Time, time.Time) {
	start := roleStart.AddDate(0, monthNum-1, 0)
	end := start.AddDate(0, 1, 0)
	return start, end
}

func getYearRange(roleStart time.Time, yearNum int) (time.Time, time.Time) {
	start := roleStart.AddDate(yearNum-1, 0, 0)
	end := start.AddDate(1, 0, 0)
	return start, end
}

// GetCurrentWeek calculates which week number we're in relative to role start
func GetCurrentWeek(roleStart time.Time) int {
	now := time.Now()
	daysSince := int(now.Sub(roleStart).Hours() / 24)
	return (daysSince / 7) + 1
}

// GetCurrentMonth calculates which month number we're in relative to role start
func GetCurrentMonth(roleStart time.Time) int {
	now := time.Now()
	daysSince := int(now.Sub(roleStart).Hours() / 24)
	return (daysSince / 30) + 1 // Approximate
}

// GetCurrentYear calculates which year number we're in relative to role start
func GetCurrentYear(roleStart time.Time) int {
	now := time.Now()
	daysSince := int(now.Sub(roleStart).Hours() / 24)
	return (daysSince / 365) + 1
}

func getTenure(roleStart time.Time) string {
	now := time.Now()

	// Calculate business days (excluding weekends and US holidays)
	businessDays := countBusinessDays(roleStart, now)

	// Calculate calendar-based weeks and months
	calendarDays := int(now.Sub(roleStart).Hours() / 24)
	weeks := calendarDays / 7
	months := int(now.Sub(roleStart).Hours() / 24 / 30.44) // Average month length

	// Calculate years as fraction
	yearsSinceStart := now.Sub(roleStart).Hours() / 24 / 365.25

	return fmt.Sprintf("Week %d | Month %d | %.2f years | %d business days since %s",
		weeks, months, yearsSinceStart, businessDays, roleStart.Format("2006-01-02"))
}

// countBusinessDays counts business days between two dates (excludes weekends and US federal holidays)
func countBusinessDays(start, end time.Time) int {
	count := 0
	current := start

	for current.Before(end) || current.Equal(end) {
		// Check if it's a weekend
		if current.Weekday() != time.Saturday && current.Weekday() != time.Sunday {
			// Check if it's a US federal holiday
			if !isUSFederalHoliday(current) {
				count++
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return count
}

// isUSFederalHoliday checks if a date is a US federal holiday
func isUSFederalHoliday(date time.Time) bool {
	month := date.Month()
	day := date.Day()

	// Fixed-date holidays
	if month == time.January && day == 1 { // New Year's Day
		return true
	}
	if month == time.July && day == 4 { // Independence Day
		return true
	}
	if month == time.November && day == 11 { // Veterans Day
		return true
	}
	if month == time.December && day == 25 { // Christmas
		return true
	}
	if month == time.June && day == 19 { // Juneteenth
		return true
	}

	// Floating holidays (Nth weekday of month)
	weekday := date.Weekday()

	// MLK Day (3rd Monday in January)
	if month == time.January && weekday == time.Monday && day >= 15 && day <= 21 {
		return true
	}

	// Presidents Day (3rd Monday in February)
	if month == time.February && weekday == time.Monday && day >= 15 && day <= 21 {
		return true
	}

	// Memorial Day (last Monday in May)
	if month == time.May && weekday == time.Monday && day >= 25 {
		return true
	}

	// Labor Day (1st Monday in September)
	if month == time.September && weekday == time.Monday && day <= 7 {
		return true
	}

	// Thanksgiving (4th Thursday in November)
	if month == time.November && weekday == time.Thursday && day >= 22 && day <= 28 {
		return true
	}

	return false
}

func parseDuration(s string) (time.Duration, error) {
	// Parse formats like "30d", "12w", "3m", "2y"
	re := regexp.MustCompile(`^(\d+)([dwmy])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format (use 30d, 12w, etc)")
	}

	// matches[1] is guaranteed to be \d+ from regex, but handle error for linter
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid number in duration: %s", matches[1])
	}
	unit := matches[2]

	switch unit {
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	case "w":
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	case "m":
		return time.Duration(num) * 30 * 24 * time.Hour, nil
	case "y":
		return time.Duration(num) * 365 * 24 * time.Hour, nil
	}

	// unit is guaranteed to be [dwmy] from regex, so this is unreachable
	return 0, fmt.Errorf("unknown unit: %s", unit)
}
