package brag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestBragDocument creates a temporary brag.json for testing
func setupTestBragDocument(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "brag.json")

	// Initialize with a basic document
	doc := &BragDocument{
		RoleTitle:     "Test Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        1,
		Entries:       []Entry{},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test document: %v", err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatalf("failed to create test brag.json: %v", err)
	}

	// Override getBragPath for tests
	originalGetBragPath := getBragPath
	getBragPath = func() (string, error) {
		return testPath, nil
	}

	cleanup := func() {
		getBragPath = originalGetBragPath
	}

	return testPath, cleanup
}

func TestSelfHealingNextID(t *testing.T) {
	// Critical: if NextID gets corrupted (manual JSON edit), should auto-fix on load
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "brag.json")

	// Create document with corrupted NextID (should be 4, but is 2)
	doc := &BragDocument{
		RoleTitle:     "Test",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        2, // WRONG - highest ID is 3
		Entries: []Entry{
			{ID: 1, Description: "first", Bucket: "Delivery", StartDate: time.Now(), EndDate: time.Now()},
			{ID: 3, Description: "third", Bucket: "Process", StartDate: time.Now(), EndDate: time.Now()}, // Gap in IDs
		},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal corrupted document: %v", err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatalf("failed to write corrupted document: %v", err)
	}

	// Override getBragPath
	originalGetBragPath := getBragPath
	getBragPath = func() (string, error) {
		return testPath, nil
	}
	defer func() { getBragPath = originalGetBragPath }()

	// Read document - should auto-heal
	loaded, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}

	if loaded.NextID != 4 {
		t.Errorf("readBragDocument() NextID = %d, want 4 (auto-healed from max ID 3)", loaded.NextID)
	}
}

func TestAddEntry_IDMonotonicity(t *testing.T) {
	// Critical: IDs must always increase, never reuse
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add first entry
	id1, err := AddEntry("Delivery", "first", "http://example.com", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	// Add second entry
	id2, err := AddEntry("Process", "second", "http://example.com", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	if id2 <= id1 {
		t.Errorf("AddEntry() id2 = %d, should be > id1 = %d", id2, id1)
	}

	// Read document and verify NextID
	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	if doc.NextID != id2+1 {
		t.Errorf("NextID = %d, want %d", doc.NextID, id2+1)
	}
}

func TestRemoveEntry_IDsNotRenumbered(t *testing.T) {
	// Critical: removing entry #2 should NOT renumber #3 to #2
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add three entries
	id1, err := AddEntry("Delivery", "first", "url", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	id2, err := AddEntry("Process", "second", "url", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	id3, err := AddEntry("Leadership", "third", "url", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	// Remove middle entry
	if err := RemoveEntry(id2); err != nil {
		t.Fatalf("RemoveEntry() error = %v", err)
	}

	// Read and verify
	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}

	if len(doc.Entries) != 2 {
		t.Errorf("After RemoveEntry(), len(Entries) = %d, want 2", len(doc.Entries))
	}

	// Verify IDs are still 1 and 3 (not renumbered to 1 and 2)
	ids := make(map[int]bool)
	for _, e := range doc.Entries {
		ids[e.ID] = true
	}

	if !ids[id1] || !ids[id3] {
		t.Errorf("After RemoveEntry(%d), remaining IDs = %v, want IDs %d and %d preserved", id2, ids, id1, id3)
	}

	if ids[id2] {
		t.Errorf("After RemoveEntry(%d), ID %d still exists", id2, id2)
	}
}

func TestUpdateEntry_PartialUpdate(t *testing.T) {
	// Critical: empty string parameters should NOT overwrite existing values
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add entry with all fields
	id, err := AddEntry("Delivery", "original description", "http://original.com", "Completed", time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), time.Date(2024, 1, 2, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	// Update only description (empty strings for other fields should preserve originals)
	if err = UpdateEntry(id, "", "updated description", "", "", time.Time{}, time.Time{}); err != nil {
		t.Fatalf("UpdateEntry() error = %v", err)
	}

	// Read and verify
	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	entry := doc.Entries[0]

	if entry.Description != "updated description" {
		t.Errorf("Description = %q, want 'updated description'", entry.Description)
	}

	if entry.Evidence != "http://original.com" {
		t.Errorf("Evidence = %q, want 'http://original.com' (should be preserved)", entry.Evidence)
	}

	if entry.Bucket != "Delivery" {
		t.Errorf("Bucket = %q, want 'Delivery' (should be preserved)", entry.Bucket)
	}
}

func TestEnrichEntry_NilVsEmptyString(t *testing.T) {
	// Critical: distinguish between "not set" (nil) and "set to empty" (skip)
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add base entry
	id, err := AddEntry("Process", "test entry", "http://example.com", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	// Enrich with hours saved - signature: (id, evidence, hoursSaved, hoursSavedCalc, businessMetric, strategicAlign, peerRecognition)
	hours := 10.0
	if err = EnrichEntry(id, "", &hours, "manual calculation", "", "improved velocity", ""); err != nil {
		t.Fatalf("EnrichEntry() error = %v", err)
	}

	// Read and verify
	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	entry := doc.Entries[0]

	if entry.HoursSaved == nil || *entry.HoursSaved != 10.0 {
		t.Errorf("HoursSaved = %v, want 10.0", entry.HoursSaved)
	}

	if entry.StrategicAlign != "improved velocity" {
		t.Errorf("StrategicAlign = %q, want 'improved velocity'", entry.StrategicAlign)
	}

	// Empty strings should result in empty fields (not nil)
	if entry.BusinessMetric != "" {
		t.Errorf("BusinessMetric = %q, want empty string", entry.BusinessMetric)
	}

	if entry.PeerRecognition != "" {
		t.Errorf("PeerRecognition = %q, want empty string", entry.PeerRecognition)
	}

	// EnrichedAt should be set
	if entry.EnrichedAt == nil {
		t.Error("EnrichedAt should be set after enrichment")
	}
}

func TestEnrichEntry_ReEnrichPreservesHours(t *testing.T) {
	// Critical: re-enriching an entry should not clear HoursSaved if nil is passed
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add and enrich
	id, err := AddEntry("Process", "test", "url", "Completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	hours := 5.0
	if err = EnrichEntry(id, "", &hours, "calc", "metric1", "", ""); err != nil {
		t.Fatalf("EnrichEntry() first call error = %v", err)
	}

	// Re-enrich without touching hours (pass nil)
	if err = EnrichEntry(id, "", nil, "", "", "new alignment", ""); err != nil {
		t.Fatalf("EnrichEntry() second call error = %v", err)
	}

	// Verify hours preserved
	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	entry := doc.Entries[0]

	if entry.HoursSaved == nil || *entry.HoursSaved != 5.0 {
		t.Errorf("After re-enrich, HoursSaved = %v, want 5.0 (preserved)", entry.HoursSaved)
	}

	if entry.StrategicAlign != "new alignment" {
		t.Errorf("StrategicAlign = %q, want 'new alignment'", entry.StrategicAlign)
	}
}

func TestClearEntries_PreservesNextID(t *testing.T) {
	// Critical: clearing entries should NOT reset NextID to 1
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Add entries
	if _, err := AddEntry("Delivery", "first", "url", "Completed", time.Now(), time.Now()); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	if _, err := AddEntry("Process", "second", "url", "Completed", time.Now(), time.Now()); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}
	if _, err := AddEntry("Leadership", "third", "url", "Completed", time.Now(), time.Now()); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	nextIDBeforeClear := doc.NextID

	// Clear
	if err := ClearEntries(); err != nil {
		t.Fatalf("ClearEntries() error = %v", err)
	}

	// Verify
	doc, err = readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}

	if len(doc.Entries) != 0 {
		t.Errorf("After ClearEntries(), len(Entries) = %d, want 0", len(doc.Entries))
	}

	if doc.NextID != nextIDBeforeClear {
		t.Errorf("After ClearEntries(), NextID = %d, want %d (preserved)", doc.NextID, nextIDBeforeClear)
	}
}

// TestCalculateCompleteness_AllScores tests the completeness scoring system
func TestCalculateCompleteness_AllScores(t *testing.T) {
	tests := []struct {
		name        string
		entry       Entry
		wantScore   int
		description string
	}{
		{
			name: "empty_entry",
			entry: Entry{
				Description:     "",
				Evidence:        "",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 0,
		},
		{
			name: "description_only",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 20,
		},
		{
			name: "evidence_only",
			entry: Entry{
				Description:     "",
				Evidence:        "http://example.com",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 20,
		},
		{
			name: "missing_evidence_marker",
			entry: Entry{
				Description:     "",
				Evidence:        "[missing]",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 0,
		},
		{
			name: "description_and_evidence",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 40,
		},
		{
			name: "with_hours_saved",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      ptrFloat(10.5),
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 55,
		},
		{
			name: "with_business_metric",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      nil,
				BusinessMetric:  "10% improvement",
				StrategicAlign:  "",
				PeerRecognition: "",
			},
			wantScore: 55,
		},
		{
			name: "with_strategic_align",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "Aligns with roadmap",
				PeerRecognition: "",
			},
			wantScore: 55,
		},
		{
			name: "with_peer_recognition",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      nil,
				BusinessMetric:  "",
				StrategicAlign:  "",
				PeerRecognition: "Great work!",
			},
			wantScore: 55,
		},
		{
			name: "fully_enriched",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      ptrFloat(10.5),
				BusinessMetric:  "10% improvement",
				StrategicAlign:  "Aligns with roadmap",
				PeerRecognition: "Great work!",
			},
			wantScore: 100,
		},
		{
			name: "three_enrichments",
			entry: Entry{
				Description:     "Some description",
				Evidence:        "http://example.com",
				HoursSaved:      ptrFloat(5.0),
				BusinessMetric:  "",
				StrategicAlign:  "Strategic",
				PeerRecognition: "Good job",
			},
			wantScore: 85, // 40 (desc+evidence) + 15 (hours) + 15 (strategic) + 15 (peer)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.CalculateCompleteness()
			if got != tt.wantScore {
				t.Errorf("CalculateCompleteness() = %d, want %d", got, tt.wantScore)
			}
		})
	}
}

// TestUpdateRoleStartDate tests role start date updates
func TestUpdateRoleStartDate(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	// Initial date
	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	originalDate := doc.RoleStartDate
	t.Logf("Original role start date: %v", originalDate)

	// Update to new date
	newDate := time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local)
	if err := UpdateRoleStartDate(newDate); err != nil {
		t.Fatalf("UpdateRoleStartDate() error = %v", err)
	}

	// Verify updated
	doc, err = readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}

	if !doc.RoleStartDate.Equal(newDate) {
		t.Errorf("RoleStartDate = %v, want %v", doc.RoleStartDate, newDate)
	}

	// Entries should be preserved
	if _, err := AddEntry("Delivery", "test", "url", "Completed", time.Now(), time.Now()); err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	doc, err = readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	if len(doc.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(doc.Entries))
	}
}

// TestUpdateRoleStartDate_NotFound tests error when brag document doesn't exist
func TestUpdateRoleStartDate_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "nonexistent", "brag.json")

	originalGetBragPath := getBragPath
	getBragPath = func() (string, error) {
		return testPath, nil
	}
	defer func() { getBragPath = originalGetBragPath }()

	newDate := time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local)
	err := UpdateRoleStartDate(newDate)
	if err == nil {
		t.Error("UpdateRoleStartDate() expected error, got nil")
	}
}

// TestInitBragDocument tests initialization (requires git/gh to succeed or fail gracefully)
func TestInitBragDocument_PathCreation(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "brag-config", "brag.json")

	originalGetBragPath := getBragPath
	getBragPath = func() (string, error) {
		return nestedPath, nil
	}
	defer func() { getBragPath = originalGetBragPath }()

	roleTitle := "Test Engineer"
	roleStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)

	err := InitBragDocument(roleTitle, roleStartDate)
	if err != nil {
		t.Logf("InitBragDocument() had expected error (git/gh may not be configured): %v", err)
		// Since git init might fail in test environment, we check if directory was created
		if _, err := os.Stat(filepath.Dir(nestedPath)); os.IsNotExist(err) {
			t.Fatalf("InitBragDocument() did not create directory structure: %v", err)
		}
		return
	}

	// If initialization succeeded, verify the file was created
	data, err := os.ReadFile(nestedPath)
	if err != nil {
		t.Fatalf("Failed to read initialized brag.json: %v", err)
	}

	var doc BragDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Failed to unmarshal initialized brag.json: %v", err)
	}

	if doc.RoleTitle != roleTitle {
		t.Errorf("RoleTitle = %q, want %q", doc.RoleTitle, roleTitle)
	}

	if !doc.RoleStartDate.Equal(roleStartDate) {
		t.Errorf("RoleStartDate = %v, want %v", doc.RoleStartDate, roleStartDate)
	}

	if doc.NextID != 1 {
		t.Errorf("NextID = %d, want 1", doc.NextID)
	}

	if len(doc.Entries) != 0 {
		t.Errorf("Entries = %d, want 0", len(doc.Entries))
	}
}

// TestInitBragDocument_AlreadyExists tests error when file already exists
func TestInitBragDocument_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "brag.json")

	// Create initial file
	doc := &BragDocument{
		RoleTitle:     "Existing Engineer",
		RoleStartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		NextID:        1,
		Entries:       []Entry{},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if err := os.WriteFile(testPath, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	originalGetBragPath := getBragPath
	getBragPath = func() (string, error) {
		return testPath, nil
	}
	defer func() { getBragPath = originalGetBragPath }()

	// Try to init again
	err = InitBragDocument("New Engineer", time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local))
	if err == nil {
		t.Error("InitBragDocument() expected error for existing file, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("InitBragDocument() error = %v, want error containing 'already exists'", err)
	}
}

// TestInitBragDocument_GetPathError tests getBragPath error handling
func TestInitBragDocument_GetPathError(t *testing.T) {
	originalGetBragPath := getBragPath
	getBragPath = func() (string, error) {
		return "", fmt.Errorf("mock getBragPath error")
	}
	defer func() { getBragPath = originalGetBragPath }()

	err := InitBragDocument("Test", time.Now())
	if err == nil {
		t.Error("InitBragDocument() expected error, got nil")
	}
}

// TestUpdateEntry_PartialFieldUpdates tests table-driven edge cases for UpdateEntry
func TestUpdateEntry_PartialFieldUpdates(t *testing.T) {
	tests := []struct {
		name         string
		bucket       string
		desc         string
		evidence     string
		status       string
		wantBucket   string
		wantDesc     string
		wantEvidence string
		wantStatus   string
	}{
		{
			name:         "update_only_bucket",
			bucket:       "Leadership",
			desc:         "",
			evidence:     "",
			status:       "",
			wantBucket:   "Leadership",
			wantDesc:     "original",
			wantEvidence: "http://original.com",
			wantStatus:   "Completed",
		},
		{
			name:         "update_only_status",
			bucket:       "",
			desc:         "",
			evidence:     "",
			status:       "In Progress",
			wantBucket:   "Delivery",
			wantDesc:     "original",
			wantEvidence: "http://original.com",
			wantStatus:   "In Progress",
		},
		{
			name:         "update_bucket_and_evidence",
			bucket:       "Process",
			desc:         "",
			evidence:     "http://new.com",
			status:       "",
			wantBucket:   "Process",
			wantDesc:     "original",
			wantEvidence: "http://new.com",
			wantStatus:   "Completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupTestBragDocument(t)
			defer cleanup()

			id, err := AddEntry("Delivery", "original", "http://original.com", "Completed", time.Now(), time.Now())
			if err != nil {
				t.Fatalf("AddEntry() error = %v", err)
			}

			if err := UpdateEntry(id, tt.bucket, tt.desc, tt.evidence, tt.status, time.Time{}, time.Time{}); err != nil {
				t.Fatalf("UpdateEntry() error = %v", err)
			}

			doc, err := readBragDocument()
			if err != nil {
				t.Fatalf("readBragDocument() error = %v", err)
			}
			entry := doc.Entries[0]

			if entry.Bucket != tt.wantBucket {
				t.Errorf("Bucket = %q, want %q", entry.Bucket, tt.wantBucket)
			}
			if entry.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", entry.Description, tt.wantDesc)
			}
			if entry.Evidence != tt.wantEvidence {
				t.Errorf("Evidence = %q, want %q", entry.Evidence, tt.wantEvidence)
			}
			if entry.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", entry.Status, tt.wantStatus)
			}
		})
	}
}

// TestUpdateEntry_DateUpdates tests date field updates in UpdateEntry
func TestUpdateEntry_DateUpdates(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	originalStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	originalEnd := time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)

	id, err := AddEntry("Delivery", "test", "url", "Completed", originalStart, originalEnd)
	if err != nil {
		t.Fatalf("AddEntry() error = %v", err)
	}

	newStart := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local)
	newEnd := time.Date(2024, 2, 28, 0, 0, 0, 0, time.Local)

	if err := UpdateEntry(id, "", "", "", "", newStart, newEnd); err != nil {
		t.Fatalf("UpdateEntry() error = %v", err)
	}

	doc, err := readBragDocument()
	if err != nil {
		t.Fatalf("readBragDocument() error = %v", err)
	}
	entry := doc.Entries[0]

	if !entry.StartDate.Equal(newStart) {
		t.Errorf("StartDate = %v, want %v", entry.StartDate, newStart)
	}
	if !entry.EndDate.Equal(newEnd) {
		t.Errorf("EndDate = %v, want %v", entry.EndDate, newEnd)
	}
}

// TestUpdateEntry_NotFound tests error for nonexistent entry
func TestUpdateEntry_NotFound(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	err := UpdateEntry(999, "Delivery", "desc", "url", "Completed", time.Time{}, time.Time{})
	if err == nil {
		t.Error("UpdateEntry() expected error for nonexistent ID, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("UpdateEntry() error = %v, want error containing 'not found'", err)
	}
}

// TestEnrichEntry_AllFieldVariations tests different enrichment combinations
func TestEnrichEntry_AllFieldVariations(t *testing.T) {
	tests := []struct {
		name                 string
		evidence             string
		hoursSaved           *float64
		hoursSavedCalc       string
		businessMetric       string
		strategicAlign       string
		peerRecognition      string
		wantEvidence         string
		wantHoursSaved       *float64
		wantHoursSavedCalc   string
		wantBusinessMetric   string
		wantStrategicAlign   string
		wantPeerRecognition  string
		wantEnrichedAtNotNil bool
	}{
		{
			name:                 "enrich_all_fields",
			evidence:             "http://new-evidence.com",
			hoursSaved:           ptrFloat(15.0),
			hoursSavedCalc:       "formula",
			businessMetric:       "5% faster",
			strategicAlign:       "roadmap",
			peerRecognition:      "nice work",
			wantEvidence:         "http://new-evidence.com",
			wantHoursSaved:       ptrFloat(15.0),
			wantHoursSavedCalc:   "formula",
			wantBusinessMetric:   "5% faster",
			wantStrategicAlign:   "roadmap",
			wantPeerRecognition:  "nice work",
			wantEnrichedAtNotNil: true,
		},
		{
			name:                 "enrich_only_metrics",
			evidence:             "",
			hoursSaved:           nil,
			hoursSavedCalc:       "",
			businessMetric:       "10%",
			strategicAlign:       "",
			peerRecognition:      "",
			wantEvidence:         "http://original.com",
			wantHoursSaved:       nil,
			wantHoursSavedCalc:   "",
			wantBusinessMetric:   "10%",
			wantStrategicAlign:   "",
			wantPeerRecognition:  "",
			wantEnrichedAtNotNil: true,
		},
		{
			name:                 "enrich_hours_and_calc_together",
			evidence:             "",
			hoursSaved:           ptrFloat(20.0),
			hoursSavedCalc:       "20 hours per month saved",
			businessMetric:       "",
			strategicAlign:       "",
			peerRecognition:      "",
			wantEvidence:         "http://original.com",
			wantHoursSaved:       ptrFloat(20.0),
			wantHoursSavedCalc:   "20 hours per month saved",
			wantBusinessMetric:   "",
			wantStrategicAlign:   "",
			wantPeerRecognition:  "",
			wantEnrichedAtNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupTestBragDocument(t)
			defer cleanup()

			id, err := AddEntry("Process", "test", "http://original.com", "Completed", time.Now(), time.Now())
			if err != nil {
				t.Fatalf("AddEntry() error = %v", err)
			}

			if err := EnrichEntry(id, tt.evidence, tt.hoursSaved, tt.hoursSavedCalc, tt.businessMetric, tt.strategicAlign, tt.peerRecognition); err != nil {
				t.Fatalf("EnrichEntry() error = %v", err)
			}

			doc, err := readBragDocument()
			if err != nil {
				t.Fatalf("readBragDocument() error = %v", err)
			}
			entry := doc.Entries[0]

			if entry.Evidence != tt.wantEvidence {
				t.Errorf("Evidence = %q, want %q", entry.Evidence, tt.wantEvidence)
			}

			if (tt.wantHoursSaved == nil && entry.HoursSaved != nil) ||
				(tt.wantHoursSaved != nil && entry.HoursSaved == nil) ||
				(tt.wantHoursSaved != nil && *entry.HoursSaved != *tt.wantHoursSaved) {
				t.Errorf("HoursSaved = %v, want %v", entry.HoursSaved, tt.wantHoursSaved)
			}

			if entry.HoursSavedCalculation != tt.wantHoursSavedCalc {
				t.Errorf("HoursSavedCalculation = %q, want %q", entry.HoursSavedCalculation, tt.wantHoursSavedCalc)
			}

			if entry.BusinessMetric != tt.wantBusinessMetric {
				t.Errorf("BusinessMetric = %q, want %q", entry.BusinessMetric, tt.wantBusinessMetric)
			}

			if entry.StrategicAlign != tt.wantStrategicAlign {
				t.Errorf("StrategicAlign = %q, want %q", entry.StrategicAlign, tt.wantStrategicAlign)
			}

			if entry.PeerRecognition != tt.wantPeerRecognition {
				t.Errorf("PeerRecognition = %q, want %q", entry.PeerRecognition, tt.wantPeerRecognition)
			}

			if tt.wantEnrichedAtNotNil && entry.EnrichedAt == nil {
				t.Error("EnrichedAt should be set")
			}
		})
	}
}

// TestEnrichEntry_NotFound tests error for nonexistent entry
func TestEnrichEntry_NotFound(t *testing.T) {
	_, cleanup := setupTestBragDocument(t)
	defer cleanup()

	hours := 10.0
	err := EnrichEntry(999, "", &hours, "", "", "", "")
	if err == nil {
		t.Error("EnrichEntry() expected error for nonexistent ID, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("EnrichEntry() error = %v, want error containing 'not found'", err)
	}
}

// Helper function to create a float pointer
func ptrFloat(f float64) *float64 {
	return &f
}
