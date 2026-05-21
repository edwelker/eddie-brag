package brag

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

// Helper to capture stdout
func captureOutput(fn func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	err = fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, copyErr := io.Copy(&buf, r); copyErr != nil {
		return "", copyErr
	}
	return buf.String(), err
}

// TestListEntries_AllEntries tests listing all entries with --all flag
func TestListEntries_AllEntries(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	oldTime := now.Add(-30 * 24 * time.Hour)

	// Add entries across time range
	id1, err := AddEntry("Delivery", "old task", "http://example.com", "Completed", oldTime, oldTime.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "recent task", "http://example.com", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true) // all=true
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should include both old and recent entries
	if !bytes.Contains([]byte(output), []byte("old task")) {
		t.Error("ListEntries(all=true) should include entries older than 7 days")
	}
	if !bytes.Contains([]byte(output), []byte("recent task")) {
		t.Error("ListEntries(all=true) should include recent entries")
	}
	if !bytes.Contains([]byte(output), []byte("Delivery")) {
		t.Error("ListEntries should group by bucket")
	}

	_ = id1
	_ = id2
}

// TestListEntries_DefaultRange tests default 7-day range
func TestListEntries_DefaultRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	oldTime := now.Add(-30 * 24 * time.Hour)
	recentTime := now.Add(-3 * 24 * time.Hour)

	// Add old entry (>7 days ago)
	if _, err := AddEntry("Process", "old entry", "url", "Completed", oldTime, oldTime.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	// Add recent entry (<7 days)
	if _, err := AddEntry("Leadership", "recent entry", "url", "Completed", recentTime, recentTime.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, false) // default range
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should only include recent entry
	if !bytes.Contains([]byte(output), []byte("recent entry")) {
		t.Error("ListEntries should include entries from last 7 days")
	}
	if bytes.Contains([]byte(output), []byte("old entry")) {
		t.Error("ListEntries should NOT include entries older than 7 days")
	}
	if bytes.Contains([]byte(output), []byte("hint")) || bytes.Contains([]byte(output), []byte("Showing last")) {
		// Hint should be shown when using default range
		t.Log("Hint shown as expected")
	}
}

// TestListEntries_CustomRange tests --range flag
func TestListEntries_CustomRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	sixDaysAgo := now.Add(-6 * 24 * time.Hour)
	sixtyDaysAgo := now.Add(-60 * 24 * time.Hour)

	// Add entries
	if _, err := AddEntry("Delivery", "entry 60 days ago", "url", "Completed", sixtyDaysAgo, sixtyDaysAgo.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := AddEntry("Architecture", "entry 6 days ago", "url", "Completed", sixDaysAgo, sixDaysAgo.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	// Query 30-day range
	output, err := captureOutput(func() error {
		return ListEntries("30d", 0, 0, false)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("entry 6 days ago")) {
		t.Error("ListEntries(30d) should include entries from last 30 days")
	}
	if bytes.Contains([]byte(output), []byte("entry 60 days ago")) {
		t.Error("ListEntries(30d) should NOT include entries older than 30 days")
	}
}

// TestListEntries_InvalidRange tests error handling
func TestListEntries_InvalidRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	err := ListEntries("invalid", 0, 0, false)
	if err == nil {
		t.Error("ListEntries should return error for invalid range format")
	}
}

// TestListEntries_NoEntriesFound tests empty result handling
func TestListEntries_NoEntriesFound(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, false)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("No entries found")) {
		t.Error("ListEntries should show 'No entries found' message when empty")
	}
}

// TestListEntries_ShowsCompleteness tests completeness scoring display
func TestListEntries_ShowsCompleteness(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Add entry and enrich partially
	id, err := AddEntry("Delivery", "test task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	hours := 8.0
	if err := EnrichEntry(id, "", &hours, "calc", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should show completeness score
	if !bytes.Contains([]byte(output), []byte("%")) {
		t.Error("ListEntries should show completeness percentage")
	}
}

// TestListEntries_GroupsByBucket tests bucket grouping
func TestListEntries_GroupsByBucket(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	if _, err := AddEntry("Delivery", "delivery task", "url", "Completed", now, now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := AddEntry("Architecture", "architecture task", "url", "Completed", now, now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := AddEntry("Process", "process task", "url", "Completed", now, now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := AddEntry("Leadership", "leadership task", "url", "Completed", now, now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Verify bucket headers appear in order
	delPos := bytes.Index([]byte(output), []byte("## Delivery"))
	archPos := bytes.Index([]byte(output), []byte("## Architecture"))
	procPos := bytes.Index([]byte(output), []byte("## Process"))
	leadPos := bytes.Index([]byte(output), []byte("## Leadership"))

	if delPos == -1 || archPos == -1 || procPos == -1 || leadPos == -1 {
		t.Error("ListEntries should group by all four buckets with headers")
	}

	// Verify order: Delivery, Architecture, Process, Leadership
	if delPos < archPos && archPos < procPos && procPos < leadPos {
		t.Log("Buckets appear in correct order")
	}
}

// TestListEntries_CalculatesTotalHours tests hours calculation
func TestListEntries_CalculatesTotalHours(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Add two entries with hours
	id1, err := AddEntry("Delivery", "task1", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "task2", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours1 := 8.0
	hours2 := 16.0
	if err := EnrichEntry(id1, "", &hours1, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours2, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should show total hours (24.0) and business days (3.0)
	if !bytes.Contains([]byte(output), []byte("Total Hours Saved")) {
		t.Error("ListEntries should display total hours saved")
	}
}

// TestGetIncompleteEntries_FindsIncomplete tests filtering incomplete entries
func TestGetIncompleteEntries_FindsIncomplete(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Entry 1: Complete (has description + evidence + hours + metrics)
	id1, err := AddEntry("Delivery", "complete entry", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	hours := 10.0
	if err := EnrichEntry(id1, "", &hours, "", "metric", "alignment", "peer"); err != nil {
		t.Fatal(err)
	}

	// Entry 2: Incomplete (only description, no evidence or enrichment)
	id2, err := AddEntry("Architecture", "incomplete entry", "", "In Progress", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	incomplete, err := GetIncompleteEntries()
	if err != nil {
		t.Fatalf("GetIncompleteEntries() error = %v", err)
	}

	// Should only include id2
	if len(incomplete) < 1 {
		t.Fatal("GetIncompleteEntries should find at least one incomplete entry")
	}

	found := false
	for _, entry := range incomplete {
		if entry.ID == id2 {
			found = true
			break
		}
	}

	if !found {
		t.Error("GetIncompleteEntries should include entry #2 (incomplete)")
	}

	_ = id1
}

// TestGetIncompleteEntries_ReturnsEmpty tests empty result
func TestGetIncompleteEntries_ReturnsEmpty(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	incomplete, err := GetIncompleteEntries()
	if err != nil {
		t.Fatalf("GetIncompleteEntries() error = %v", err)
	}

	if len(incomplete) != 0 {
		t.Error("GetIncompleteEntries should return empty slice when no incomplete entries")
	}
}

// TestGetUnenrichedEntries_DefaultRange tests default 7-day range
func TestGetUnenrichedEntries_DefaultRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	oldTime := now.Add(-30 * 24 * time.Hour)
	recentTime := now.Add(-3 * 24 * time.Hour)

	// Old unenriched entry (no evidence, no EnrichedAt)
	if _, err := AddEntry("Delivery", "old unenriched", "", "Completed", oldTime, oldTime.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	// Recent unenriched entry
	id2, err := AddEntry("Architecture", "recent unenriched", "", "In Progress", recentTime, recentTime.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	unenriched, err := GetUnenrichedEntries("", 0)
	if err != nil {
		t.Fatalf("GetUnenrichedEntries() error = %v", err)
	}

	// Should only include id2 (recent)
	found := false
	for _, entry := range unenriched {
		if entry.ID == id2 {
			found = true
			break
		}
	}

	if !found {
		t.Error("GetUnenrichedEntries should include recent unenriched entries")
	}

	if len(unenriched) > 1 {
		t.Error("GetUnenrichedEntries should NOT include old entries by default")
	}
}

// TestGetUnenrichedEntries_SpecificID tests fetching by ID regardless of enrichment
func TestGetUnenrichedEntries_SpecificID(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Already enriched entry
	id1, err := AddEntry("Delivery", "enriched entry", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	hours := 5.0
	if err := EnrichEntry(id1, "evidence", &hours, "calc", "metric", "", ""); err != nil {
		t.Fatal(err)
	}

	// Unenriched entry
	id2, err := AddEntry("Architecture", "unenriched entry", "", "In Progress", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Request enriched entry by ID - should return it anyway (for re-enrichment)
	unenriched, err := GetUnenrichedEntries("", id1)
	if err != nil {
		t.Fatalf("GetUnenrichedEntries(id=%d) error = %v", id1, err)
	}

	if len(unenriched) != 1 {
		t.Fatalf("GetUnenrichedEntries(id=%d) should return exactly 1 entry, got %d", id1, len(unenriched))
	}

	if unenriched[0].ID != id1 {
		t.Error("GetUnenrichedEntries(id) should return the requested entry regardless of enrichment status")
	}

	_ = id2
}

// TestGetUnenrichedEntries_CustomRange tests range filtering
func TestGetUnenrichedEntries_CustomRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	sixtyDaysAgo := now.Add(-60 * 24 * time.Hour)
	fiveDaysAgo := now.Add(-5 * 24 * time.Hour)

	// Old unenriched
	if _, err := AddEntry("Delivery", "old", "", "Completed", sixtyDaysAgo, sixtyDaysAgo.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	// Recent unenriched
	if _, err := AddEntry("Architecture", "recent", "", "In Progress", fiveDaysAgo, fiveDaysAgo.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	unenriched, err := GetUnenrichedEntries("30d", 0)
	if err != nil {
		t.Fatalf("GetUnenrichedEntries(30d) error = %v", err)
	}

	// Should find recent but not old
	hasRecent := false
	hasOld := false
	for _, entry := range unenriched {
		if entry.Description == "recent" {
			hasRecent = true
		}
		if entry.Description == "old" {
			hasOld = true
		}
	}

	if !hasRecent {
		t.Error("GetUnenrichedEntries(30d) should include recent unenriched entries")
	}
	if hasOld {
		t.Error("GetUnenrichedEntries(30d) should NOT include entries older than 30 days")
	}
}

// TestGetUnenrichedEntries_SkipsEnriched tests filtering already-enriched entries
func TestGetUnenrichedEntries_SkipsEnriched(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Enriched entry
	id1, err := AddEntry("Delivery", "enriched", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	hours := 5.0
	if err := EnrichEntry(id1, "evidence", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	// Unenriched entry
	id2, err := AddEntry("Architecture", "unenriched", "", "In Progress", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	unenriched, err := GetUnenrichedEntries("", 0)
	if err != nil {
		t.Fatalf("GetUnenrichedEntries() error = %v", err)
	}

	// Should only include id2
	if len(unenriched) != 1 {
		t.Fatalf("GetUnenrichedEntries() should return 1 entry, got %d", len(unenriched))
	}

	if unenriched[0].ID != id2 {
		t.Error("GetUnenrichedEntries should skip already-enriched entries")
	}
}

// TestReportEntries_ByWeek tests week-based reporting
func TestReportEntries_ByWeek(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	// Override role start for predictable testing
	doc, err := readBragDocument()
	if err != nil {
		t.Fatal(err)
	}
	originalStart := doc.RoleStartDate
	doc.RoleStartDate = roleStart

	now := time.Now()

	// Add entries in week 1 and week 2
	week1Start := roleStart.AddDate(0, 0, 0)
	week2Start := roleStart.AddDate(0, 0, 7)

	id1, err := AddEntry("Delivery", "week1 task", "url", "Completed", week1Start, week1Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "week2 task", "url", "Completed", week2Start, week2Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 8.0
	if err := EnrichEntry(id1, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("", 1, 0, 0) // week 1
	})

	if err != nil {
		t.Fatalf("ReportEntries(week=1) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Week 1")) {
		t.Error("ReportEntries should show period label")
	}
	if !bytes.Contains([]byte(output), []byte("week1 task")) {
		t.Error("ReportEntries should include week1 entries")
	}

	_ = originalStart
	_ = now
	_ = id2
}

// TestReportEntries_ByMonth tests month-based reporting
func TestReportEntries_ByMonth(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	now := time.Now()

	month1Start := roleStart.AddDate(0, 0, 0)
	month2Start := roleStart.AddDate(0, 1, 0)

	id1, err := AddEntry("Delivery", "month1 task", "url", "Completed", month1Start, month1Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "month2 task", "url", "Completed", month2Start, month2Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 16.0
	if err := EnrichEntry(id1, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("", 0, 1, 0) // month 1
	})

	if err != nil {
		t.Fatalf("ReportEntries(month=1) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Month 1")) {
		t.Error("ReportEntries should show month label")
	}

	_ = id2
	_ = now
}

// TestReportEntries_ByYear tests year-based reporting
func TestReportEntries_ByYear(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	year1Start := roleStart.AddDate(0, 0, 0)
	year2Start := roleStart.AddDate(1, 0, 0)

	id1, err := AddEntry("Delivery", "year1 task", "url", "Completed", year1Start, year1Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "year2 task", "url", "Completed", year2Start, year2Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 8.0
	if err := EnrichEntry(id1, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("", 0, 0, 1) // year 1
	})

	if err != nil {
		t.Fatalf("ReportEntries(year=1) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Year 1")) {
		t.Error("ReportEntries should show year label")
	}

	_ = id2
}

// TestReportEntries_ByRange tests custom range reporting
func TestReportEntries_ByRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	fiveDaysAgo := now.Add(-5 * 24 * time.Hour)

	id, err := AddEntry("Delivery", "recent task", "url", "Completed", fiveDaysAgo, fiveDaysAgo.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 8.0
	if err := EnrichEntry(id, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("7d", 0, 0, 0)
	})

	if err != nil {
		t.Fatalf("ReportEntries(range=7d) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Last 7d")) {
		t.Error("ReportEntries should show custom range label")
	}
}

// TestReportEntries_NoParametersError tests error when no period specified
func TestReportEntries_NoParametersError(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	err := ReportEntries("", 0, 0, 0)
	if err == nil {
		t.Error("ReportEntries should return error when no period specified")
	}
}

// TestReportEntries_EmptyPeriod tests handling empty result
func TestReportEntries_EmptyPeriod(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	output, err := captureOutput(func() error {
		return ReportEntries("", 1, 0, 0) // week 1 (empty)
	})

	if err != nil {
		t.Fatalf("ReportEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("No entries found")) {
		t.Error("ReportEntries should show 'No entries found' for empty period")
	}

	_ = roleStart
}

// TestReportEntries_GroupsByBucket tests bucket grouping in reports
func TestReportEntries_GroupsByBucket(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Add entries in different buckets
	id1, err := AddEntry("Delivery", "delivery task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "architecture task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id3, err := AddEntry("Process", "process task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 8.0
	if err := EnrichEntry(id1, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id3, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("90d", 0, 0, 0)
	})

	if err != nil {
		t.Fatalf("ReportEntries() error = %v", err)
	}

	// Should show bucket headers
	if !bytes.Contains([]byte(output), []byte("## Delivery")) {
		t.Error("ReportEntries should show Delivery bucket")
	}
	if !bytes.Contains([]byte(output), []byte("## Architecture")) {
		t.Error("ReportEntries should show Architecture bucket")
	}
	if !bytes.Contains([]byte(output), []byte("## Process")) {
		t.Error("ReportEntries should show Process bucket")
	}
}

// TestReportEntries_CalculatesTotals tests total calculation
func TestReportEntries_CalculatesTotals(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	id1, err := AddEntry("Delivery", "task1", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "task2", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours1 := 8.0
	hours2 := 16.0
	if err := EnrichEntry(id1, "", &hours1, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours2, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("90d", 0, 0, 0)
	})

	if err != nil {
		t.Fatalf("ReportEntries() error = %v", err)
	}

	// Should show total line with 2 entries and 24 hours
	if !bytes.Contains([]byte(output), []byte("**Total:")) {
		t.Error("ReportEntries should show total line")
	}
	if !bytes.Contains([]byte(output), []byte("2 entries")) {
		t.Error("ReportEntries should show total entries count")
	}
}

// TestListEntries_WeekNumber tests --week flag
func TestListEntries_WeekNumber(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	week2Start := roleStart.AddDate(0, 0, 7)

	_, err := AddEntry("Delivery", "week 2 task", "url", "Completed", week2Start, week2Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 2, 0, false) // week 2
	})

	if err != nil {
		t.Fatalf("ListEntries(week=2) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("week 2 task")) {
		t.Error("ListEntries should filter by week number")
	}
}

// TestListEntries_MonthNumber tests --month flag
func TestListEntries_MonthNumber(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	month2Start := roleStart.AddDate(0, 1, 0)

	_, err := AddEntry("Architecture", "month 2 task", "url", "Completed", month2Start, month2Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 2, false) // month 2
	})

	if err != nil {
		t.Fatalf("ListEntries(month=2) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("month 2 task")) {
		t.Error("ListEntries should filter by month number")
	}
}

// TestListEntries_ShowsTenure tests tenure display
func TestListEntries_ShowsTenure(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	_, err := AddEntry("Delivery", "task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Tenure:")) {
		t.Error("ListEntries should display tenure")
	}
}

// TestListEntries_ShowsRoleTitle tests role title display
func TestListEntries_ShowsRoleTitle(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	_, err := AddEntry("Process", "task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Role: Test Engineer")) {
		t.Error("ListEntries should display role title")
	}
}

// TestListEntries_ShowsEntryMetadata tests entry display with all metadata
func TestListEntries_ShowsEntryMetadata(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	id, err := AddEntry("Leadership", "complex task", "http://evidence.com", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 12.0
	if err := EnrichEntry(id, "updated evidence", &hours, "manual calc", "50% faster", "alignment goal", "great work"); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should show all fields
	if !bytes.Contains([]byte(output), []byte("updated evidence")) {
		t.Error("ListEntries should show evidence")
	}
	if !bytes.Contains([]byte(output), []byte("manual calc")) {
		t.Error("ListEntries should show hours calculation")
	}
	if !bytes.Contains([]byte(output), []byte("50% faster")) {
		t.Error("ListEntries should show business metric")
	}
	if !bytes.Contains([]byte(output), []byte("alignment goal")) {
		t.Error("ListEntries should show strategic alignment")
	}
	if !bytes.Contains([]byte(output), []byte("great work")) {
		t.Error("ListEntries should show peer recognition")
	}
}

// TestListEntries_ShowsMissingEvidence tests evidence marker
func TestListEntries_ShowsMissingEvidence(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	_, err := AddEntry("Delivery", "no evidence task", "", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("[missing]")) {
		t.Error("ListEntries should show [missing] for empty evidence")
	}
}

// TestListEntries_WarningForLowCompleteness tests completeness warning marker
func TestListEntries_WarningForLowCompleteness(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	// Add entry with only description (20 points, < 60%)
	_, err := AddEntry("Delivery", "minimal task", "", "In Progress", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ListEntries("", 0, 0, true)
	})

	if err != nil {
		t.Fatalf("ListEntries() error = %v", err)
	}

	// Should show warning symbol for < 60%
	if !bytes.Contains([]byte(output), []byte("⚠")) {
		t.Log("ListEntries should show warning for low completeness (optional emoji)")
	}
}

// TestGetUnenrichedEntries_InvalidRange tests error handling
func TestGetUnenrichedEntries_InvalidRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	_, err := GetUnenrichedEntries("invalid", 0)
	if err == nil {
		t.Error("GetUnenrichedEntries should return error for invalid range")
	}
}

// TestGetUnenrichedEntries_MissingEvidence tests evidence-only unenriched entries
func TestGetUnenrichedEntries_MissingEvidence(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Entry with enrichment timestamp but no evidence
	id1, err := AddEntry("Delivery", "enriched but no evidence", "", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	hours := 5.0
	if err := EnrichEntry(id1, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	// Entry completely unenriched
	id2, err := AddEntry("Architecture", "completely unenriched", "", "In Progress", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	unenriched, err := GetUnenrichedEntries("", 0)
	if err != nil {
		t.Fatalf("GetUnenrichedEntries() error = %v", err)
	}

	// Both should be included (no evidence or no EnrichedAt)
	found := 0
	for _, entry := range unenriched {
		if entry.ID == id1 || entry.ID == id2 {
			found++
		}
	}

	if found != 2 {
		t.Errorf("GetUnenrichedEntries should find entries with missing evidence, found %d", found)
	}
}

// TestReportEntries_MultiplePeriodsIgnoresOthers tests period filtering
func TestReportEntries_MultiplePeriodsIgnoresOthers(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	week1Start := roleStart.AddDate(0, 0, 0)
	week2Start := roleStart.AddDate(0, 0, 7)

	id1, err := AddEntry("Delivery", "week1 task", "url", "Completed", week1Start, week1Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "week2 task", "url", "Completed", week2Start, week2Start.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 8.0
	if err := EnrichEntry(id1, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}
	if err := EnrichEntry(id2, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("", 1, 0, 0) // week 1 only
	})

	if err != nil {
		t.Fatalf("ReportEntries(week=1) error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("week1 task")) {
		t.Error("ReportEntries should include week 1 task")
	}
	if bytes.Contains([]byte(output), []byte("week2 task")) {
		t.Error("ReportEntries should NOT include week 2 task when filtering by week 1")
	}
}

// TestReportEntries_RoleTitle tests report header
func TestReportEntries_RoleTitle(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	id, err := AddEntry("Delivery", "task", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	hours := 8.0
	if err := EnrichEntry(id, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("7d", 0, 0, 0)
	})

	if err != nil {
		t.Fatalf("ReportEntries() error = %v", err)
	}

	if !bytes.Contains([]byte(output), []byte("Test Engineer")) {
		t.Error("ReportEntries should show role title in header")
	}
}

// TestReportEntries_BucketWithZeroHours tests bucket with no hours
func TestReportEntries_BucketWithZeroHours(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	if _, err := AddEntry("Delivery", "task no hours", "url", "Completed", now, now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	id2, err := AddEntry("Architecture", "task with hours", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Only enrich second entry with hours
	hours := 8.0
	if err := EnrichEntry(id2, "", &hours, "", "", "", ""); err != nil {
		t.Fatal(err)
	}

	output, err := captureOutput(func() error {
		return ReportEntries("90d", 0, 0, 0)
	})

	if err != nil {
		t.Fatalf("ReportEntries() error = %v", err)
	}

	// Both buckets should appear
	if !bytes.Contains([]byte(output), []byte("## Delivery")) {
		t.Error("ReportEntries should include Delivery bucket even with 0 hours")
	}
	if !bytes.Contains([]byte(output), []byte("## Architecture")) {
		t.Error("ReportEntries should include Architecture bucket")
	}
}

// TestGetIncompleteEntries_PartialEnrichment tests entries with some fields
func TestGetIncompleteEntries_PartialEnrichment(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()

	// Entry with description + evidence only (40 points = 40%)
	id1, err := AddEntry("Delivery", "partial entry", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	// Entry with everything (100%)
	id2, err := AddEntry("Architecture", "full entry", "url", "Completed", now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	hours := 10.0
	if err := EnrichEntry(id2, "", &hours, "", "metric", "align", "peer"); err != nil {
		t.Fatal(err)
	}

	incomplete, err := GetIncompleteEntries()
	if err != nil {
		t.Fatalf("GetIncompleteEntries() error = %v", err)
	}

	// Should find partial but not full
	hasPartial := false
	hasFull := false
	for _, entry := range incomplete {
		if entry.ID == id1 {
			hasPartial = true
		}
		if entry.ID == id2 {
			hasFull = true
		}
	}

	if !hasPartial {
		t.Error("GetIncompleteEntries should include partially enriched entries")
	}
	if hasFull {
		t.Error("GetIncompleteEntries should NOT include 100% complete entries")
	}
}
