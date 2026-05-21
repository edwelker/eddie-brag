package brag

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestExportEntries_JSON tests JSON export format
func TestExportEntries_JSON(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add a test entry (use recent date to match default 7-day filter)
	now := time.Now()
	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        2,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "test entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   now,
				EndDate:     now.AddDate(0, 0, 1),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Get export dir
	exportDir := filepath.Dir(testPath)
	exportedFile := filepath.Join(exportDir, "brag.json")

	if err := ExportEntries("json", "", 0, 0, false); err != nil {
		t.Fatalf("ExportEntries(json) error = %v", err)
	}

	// Verify file exists
	if stat, err := os.Stat(exportedFile); err != nil {
		t.Errorf("ExportEntries(json) did not create exported file at %s: %v", exportedFile, err)
	} else if stat.Size() == 0 {
		t.Errorf("ExportEntries(json) created empty file at %s", exportedFile)
	}
}

// TestExportEntries_CSV tests CSV export format
func TestExportEntries_CSV(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add a test entry (use recent date to match default 7-day filter)
	now := time.Now()
	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        2,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "test entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   now,
				EndDate:     now.AddDate(0, 0, 1),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Get export dir
	exportDir := filepath.Dir(testPath)
	expectedFile := filepath.Join(exportDir, "brag.csv")

	// Clean up any previous exports
	os.Remove(expectedFile)

	if err := ExportEntries("csv", "", 0, 0, false); err != nil {
		t.Fatalf("ExportEntries(csv) error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(expectedFile); err != nil {
		t.Errorf("ExportEntries(csv) did not create expected file: %v", err)
	}
}

// TestExportEntries_TXT tests TXT export format
func TestExportEntries_TXT(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add a test entry (use recent date to match default 7-day filter)
	now := time.Now()
	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        2,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "test entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   now,
				EndDate:     now.AddDate(0, 0, 1),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Get export dir
	exportDir := filepath.Dir(testPath)
	expectedFile := filepath.Join(exportDir, "brag.txt")

	// Clean up any previous exports
	os.Remove(expectedFile)

	if err := ExportEntries("txt", "", 0, 0, false); err != nil {
		t.Fatalf("ExportEntries(txt) error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(expectedFile); err != nil {
		t.Errorf("ExportEntries(txt) did not create expected file: %v", err)
	}
}

// TestExportEntries_UnsupportedFormat tests error handling for unsupported format
func TestExportEntries_UnsupportedFormat(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	err := ExportEntries("pdf", "", 0, 0, false)
	if err == nil {
		t.Error("ExportEntries() with unsupported format should error")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("ExportEntries() error = %q, want to contain 'unsupported format'", err.Error())
	}
}

// TestExportEntries_FilterAll tests filtering with all=true flag
func TestExportEntries_FilterAll(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Create document with old and new entries
	oldDate := time.Now().AddDate(0, 0, -30)
	newDate := time.Now()

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        3,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "old entry",
				Evidence:    "http://old.com",
				Status:      "Completed",
				StartDate:   oldDate,
				EndDate:     oldDate,
			},
			{
				ID:          2,
				Bucket:      "Process",
				Description: "new entry",
				Evidence:    "http://new.com",
				Status:      "Completed",
				StartDate:   newDate,
				EndDate:     newDate,
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export with all=true
	if err := ExportEntries("json", "", 0, 0, true); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	// Read and verify both entries exported
	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.json")
	data, err = os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported []Entry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}

	if len(exported) != 2 {
		t.Errorf("ExportEntries(all=true) exported %d entries, want 2", len(exported))
	}

	// Find entries by ID
	idMap := make(map[int]Entry)
	for _, e := range exported {
		idMap[e.ID] = e
	}

	if _, ok := idMap[1]; !ok {
		t.Error("ExportEntries(all=true) did not include old entry (ID=1)")
	}
	if _, ok := idMap[2]; !ok {
		t.Error("ExportEntries(all=true) did not include new entry (ID=2)")
	}
}

// TestExportEntries_FilterByWeek tests filtering by week number
func TestExportEntries_FilterByWeek(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: roleStart,
		NextID:        3,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "week 1 entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 3, 0, 0, 0, 0, time.Local),
			},
			{
				ID:          2,
				Bucket:      "Process",
				Description: "week 2 entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 8, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export week 1
	if err := ExportEntries("json", "", 1, 0, false); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.json")
	data, err = os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported []Entry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}

	if len(exported) != 1 {
		t.Errorf("ExportEntries(week=1) exported %d entries, want 1", len(exported))
	}
	if len(exported) > 0 && exported[0].ID != 1 {
		t.Errorf("ExportEntries(week=1) exported ID=%d, want ID=1", exported[0].ID)
	}
}

// TestExportEntries_FilterByMonth tests filtering by month number
func TestExportEntries_FilterByMonth(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	roleStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: roleStart,
		NextID:        3,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "january entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local),
			},
			{
				ID:          2,
				Bucket:      "Process",
				Description: "february entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 2, 15, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 2, 20, 0, 0, 0, 0, time.Local),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export month 1
	if err := ExportEntries("json", "", 0, 1, false); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.json")
	data, err = os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported []Entry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}

	if len(exported) != 1 {
		t.Errorf("ExportEntries(month=1) exported %d entries, want 1", len(exported))
	}
	if len(exported) > 0 && exported[0].ID != 1 {
		t.Errorf("ExportEntries(month=1) exported ID=%d, want ID=1", exported[0].ID)
	}
}

// TestExportEntries_FilterByRange tests filtering by range string (e.g., "7d", "2w")
func TestExportEntries_FilterByRange(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	oldDate := now.AddDate(0, 0, -30)
	recentDate := now.AddDate(0, 0, -5)

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        3,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "old entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   oldDate,
				EndDate:     oldDate,
			},
			{
				ID:          2,
				Bucket:      "Process",
				Description: "recent entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   recentDate,
				EndDate:     recentDate,
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export last 7 days
	if err := ExportEntries("json", "7d", 0, 0, false); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.json")
	data, err = os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported []Entry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}

	// Should only have recent entry
	if len(exported) != 1 {
		t.Errorf("ExportEntries(range=7d) exported %d entries, want 1", len(exported))
	}
	if len(exported) > 0 && exported[0].ID != 2 {
		t.Errorf("ExportEntries(range=7d) exported ID=%d, want ID=2", exported[0].ID)
	}
}

// TestExportEntries_InvalidRange tests error handling for invalid range
func TestExportEntries_InvalidRange(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	err := ExportEntries("json", "invalid", 0, 0, false)
	if err == nil {
		t.Error("ExportEntries() with invalid range should error")
	}
	if !strings.Contains(err.Error(), "invalid range") {
		t.Errorf("ExportEntries() error = %q, want to contain 'invalid range'", err.Error())
	}
}

// TestExportEntries_DefaultFilter tests default 7-day filter when no flags provided
func TestExportEntries_DefaultFilter(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	now := time.Now()
	oldDate := now.AddDate(0, 0, -14)
	recentDate := now.AddDate(0, 0, -3)

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        3,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "very old entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   oldDate,
				EndDate:     oldDate,
			},
			{
				ID:          2,
				Bucket:      "Process",
				Description: "recent entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   recentDate,
				EndDate:     recentDate,
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export with no filters (should default to 7 days)
	if err := ExportEntries("json", "", 0, 0, false); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.json")
	data, err = os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported []Entry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}

	// Should only have recent entry
	if len(exported) != 1 {
		t.Errorf("ExportEntries() with default filter exported %d entries, want 1", len(exported))
	}
	if len(exported) > 0 && exported[0].ID != 2 {
		t.Errorf("ExportEntries() with default filter exported ID=%d, want ID=2", exported[0].ID)
	}
}

// TestExportEntries_SortedByBucket tests that exported entries are sorted by bucket
func TestExportEntries_SortedByBucket(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        4,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Leadership",
				Description: "leadership entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			},
			{
				ID:          2,
				Bucket:      "Delivery",
				Description: "delivery entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			},
			{
				ID:          3,
				Bucket:      "Process",
				Description: "process entry",
				Evidence:    "http://example.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export all
	if err := ExportEntries("json", "", 0, 0, true); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.json")
	data, err = os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	var exported []Entry
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}

	// Verify sorted by bucket (Delivery < Leadership < Process)
	for i := 0; i < len(exported)-1; i++ {
		if exported[i].Bucket > exported[i+1].Bucket {
			t.Errorf("ExportEntries() entries not sorted by bucket: %v > %v", exported[i].Bucket, exported[i+1].Bucket)
		}
	}
}

// TestExportEntries_EmptyDocument tests exporting an empty document
func TestExportEntries_EmptyDocument(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Document is already empty from setupTestBragDocument
	if err := ExportEntries("csv", "", 0, 0, true); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.csv")
	_, err := os.Stat(exportFile)
	if err != nil {
		t.Errorf("ExportEntries() failed to create CSV for empty document: %v", err)
	}

	// Verify CSV header is present but no data rows
	data, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 || !strings.Contains(lines[0], "ID,StartDate") {
		t.Error("ExportEntries() CSV missing header")
	}
}

// TestExportEntries_CSV_Format tests CSV export format details
func TestExportEntries_CSV_Format(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	hours := 5.5
	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        2,
		Entries: []Entry{
			{
				ID:              1,
				Bucket:          "Delivery",
				Description:     "test, with comma",
				Evidence:        "http://example.com",
				Status:          "Completed",
				HoursSaved:      &hours,
				BusinessMetric:  "saved time",
				StrategicAlign:  "aligned",
				PeerRecognition: "praised",
				StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndDate:         time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export to CSV
	if err := ExportEntries("csv", "", 0, 0, true); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.csv")
	csvData, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	// Parse CSV to verify format
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) < 2 {
		t.Error("CSV should have header + at least 1 data row")
		return
	}

	// Verify header
	expectedHeaders := []string{"ID", "StartDate", "EndDate", "Bucket", "Status", "Description", "Evidence", "HoursSaved", "BusinessMetric", "StrategicAlignment", "PeerRecognition"}
	for i, header := range expectedHeaders {
		if i >= len(records[0]) || records[0][i] != header {
			t.Errorf("CSV header[%d] = %q, want %q", i, records[0][i], header)
		}
	}

	// Verify data row
	dataRow := records[1]
	if dataRow[0] != "1" {
		t.Errorf("CSV ID = %s, want 1", dataRow[0])
	}
	if dataRow[3] != "Delivery" {
		t.Errorf("CSV Bucket = %s, want Delivery", dataRow[3])
	}
	if dataRow[7] != "5.5" {
		t.Errorf("CSV HoursSaved = %s, want 5.5", dataRow[7])
	}
}

// TestExportEntries_TXT_Format tests TXT export format and grouping
func TestExportEntries_TXT_Format(t *testing.T) {
	testPath, cleanup := setupTestBragDocument(t)
	defer cleanup()

	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        3,
		Entries: []Entry{
			{
				ID:          1,
				Bucket:      "Delivery",
				Description: "first delivery item",
				Evidence:    "http://url1.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local),
			},
			{
				ID:          2,
				Bucket:      "Delivery",
				Description: "second delivery item",
				Evidence:    "http://url2.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 4, 0, 0, 0, 0, time.Local),
			},
			{
				ID:          3,
				Bucket:      "Process",
				Description: "process item",
				Evidence:    "http://url3.com",
				Status:      "Completed",
				StartDate:   time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local),
				EndDate:     time.Date(2024, 1, 6, 0, 0, 0, 0, time.Local),
			},
		},
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Export to TXT
	if err := ExportEntries("txt", "", 0, 0, true); err != nil {
		t.Fatalf("ExportEntries() error = %v", err)
	}

	exportDir := filepath.Dir(testPath)
	exportFile := filepath.Join(exportDir, "brag.txt")
	txtData, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}

	content := string(txtData)

	// Verify bucket grouping
	if !strings.Contains(content, "## Delivery") {
		t.Error("TXT missing Delivery bucket header")
	}
	if !strings.Contains(content, "## Process") {
		t.Error("TXT missing Process bucket header")
	}

	// Verify entries appear in correct buckets
	if !strings.Contains(content, "first delivery item") {
		t.Error("TXT missing first delivery item")
	}
	if !strings.Contains(content, "second delivery item") {
		t.Error("TXT missing second delivery item")
	}
	if !strings.Contains(content, "process item") {
		t.Error("TXT missing process item")
	}

	// Verify entry format
	if !strings.Contains(content, "#1 [Completed]") {
		t.Error("TXT missing proper entry format (#1 [Completed])")
	}
}

func TestExportJSON(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		path    string
		wantErr bool
		wantIDs []int
	}{
		{
			name: "single entry",
			entries: []Entry{
				{ID: 1, Description: "test", Bucket: "Delivery", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local)},
			},
			path:    "",
			wantErr: false,
			wantIDs: []int{1},
		},
		{
			name: "multiple entries",
			entries: []Entry{
				{ID: 1, Description: "test1", Bucket: "Delivery"},
				{ID: 2, Description: "test2", Bucket: "Process"},
			},
			path:    "",
			wantErr: false,
			wantIDs: []int{1, 2},
		},
		{
			name:    "empty entries",
			entries: []Entry{},
			path:    "",
			wantErr: false,
			wantIDs: []int{},
		},
		{
			name:    "write to read-only directory",
			entries: []Entry{{ID: 1}},
			path:    "/dev/null/test.json",
			wantErr: true,
			wantIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.path != "" {
				path = tt.path
			} else {
				tmpDir := t.TempDir()
				path = filepath.Join(tmpDir, "test.json")
			}

			err := exportJSON(path, tt.entries)
			if (err != nil) != tt.wantErr {
				t.Errorf("exportJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// Read and verify
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}

			var got []Entry
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			if len(got) != len(tt.wantIDs) {
				t.Errorf("exportJSON() wrote %d entries, want %d", len(got), len(tt.wantIDs))
			}

			for i, wantID := range tt.wantIDs {
				if i >= len(got) {
					break
				}
				if got[i].ID != wantID {
					t.Errorf("entry[%d].ID = %d, want %d", i, got[i].ID, wantID)
				}
			}
		})
	}
}

func TestExportJSON_WriteError(t *testing.T) {
	// Test write to an impossible path (inside /dev/null)
	// This tests the os.WriteFile error path on line 72
	entries := []Entry{
		{ID: 1, Description: "test", Bucket: "Delivery", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local)},
	}

	err := exportJSON("/dev/null/test.json", entries)
	if err == nil {
		t.Fatal("exportJSON() expected error writing to /dev/null/test.json, got nil")
	}
}

func TestExportCSV(t *testing.T) {
	hours := 10.5
	tests := []struct {
		name         string
		entries      []Entry
		wantContains []string
	}{
		{
			name: "single entry",
			entries: []Entry{
				{ID: 1, Description: "test", Evidence: "http://example.com", Bucket: "Delivery", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local)},
			},
			wantContains: []string{"ID,StartDate", "1,2024-01-01", "test"},
		},
		{
			name: "description with comma",
			entries: []Entry{
				{ID: 2, Description: "test, with comma", Bucket: "Process", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
			},
			wantContains: []string{"\"test, with comma\""},
		},
		{
			name: "description with quotes",
			entries: []Entry{
				{ID: 3, Description: "test \"quoted\" text", Bucket: "Leadership", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
			},
			wantContains: []string{"test \"\"quoted\"\" text"},
		},
		{
			name: "entry with hours saved",
			entries: []Entry{
				{ID: 4, Description: "test", HoursSaved: &hours, Bucket: "Process", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
			},
			wantContains: []string{"10.5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.csv")

			if err := exportCSV(path, tt.entries); err != nil {
				t.Fatalf("exportCSV() error = %v", err)
			}

			// Read and verify
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}

			content := string(data)
			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("exportCSV() output missing substring %q\nGot:\n%s", want, content)
				}
			}
		})
	}
}

func TestExportTXT(t *testing.T) {
	tests := []struct {
		name         string
		entries      []Entry
		wantContains []string
	}{
		{
			name: "single entry",
			entries: []Entry{
				{ID: 1, Description: "test entry", Evidence: "http://example.com", Bucket: "Delivery", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local)},
			},
			wantContains: []string{"## Delivery", "#1", "test entry", "Evidence: http://example.com"},
		},
		{
			name: "grouped by bucket",
			entries: []Entry{
				{ID: 1, Description: "delivery item", Bucket: "Delivery", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
				{ID: 2, Description: "process item", Bucket: "Process", Status: "Completed", StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
			},
			wantContains: []string{"## Delivery", "delivery item", "## Process", "process item"},
		},
		{
			name: "enrichment fields included",
			entries: []Entry{
				{
					ID:              1,
					Description:     "test",
					Evidence:        "url",
					Bucket:          "Leadership",
					Status:          "Completed",
					BusinessMetric:  "saved time",
					StrategicAlign:  "aligned with goals",
					PeerRecognition: "praised by team",
					StartDate:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
					EndDate:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
				},
			},
			wantContains: []string{"Business Metric: saved time", "Strategic Alignment: aligned with goals", "Peer Recognition: praised by team"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.txt")

			if err := exportTXT(path, tt.entries); err != nil {
				t.Fatalf("exportTXT() error = %v", err)
			}

			// Read and verify
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}

			content := string(data)
			for _, want := range tt.wantContains {
				if !strings.Contains(content, want) {
					t.Errorf("exportTXT() output missing substring %q\nGot:\n%s", want, content)
				}
			}
		})
	}
}
