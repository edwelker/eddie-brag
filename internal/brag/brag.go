package brag

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents a single accomplishment entry
type Entry struct {
	ID                    int        `json:"id"`
	StartDate             time.Time  `json:"start_date"`
	EndDate               time.Time  `json:"end_date"`
	Bucket                string     `json:"bucket"`
	Description           string     `json:"description"`
	Evidence              string     `json:"evidence"`
	Status                string     `json:"status"`
	HoursSaved            *float64   `json:"hours_saved,omitempty"`
	HoursSavedCalculation string     `json:"hours_saved_calculation,omitempty"`
	TeamSize              *int       `json:"team_size,omitempty"`
	BusinessMetric        string     `json:"business_metric,omitempty"`
	StrategicAlign        string     `json:"strategic_alignment,omitempty"`
	PeerRecognition       string     `json:"peer_recognition,omitempty"`
	EnrichedAt            *time.Time `json:"enriched_at,omitempty"`
}

// BragDocument represents the entire brag document
type BragDocument struct {
	RoleTitle     string    `json:"role_title"`
	RoleStartDate time.Time `json:"role_start_date"`
	NextID        int       `json:"next_id"`
	Entries       []Entry   `json:"entries"`
}

// CalculateCompleteness returns a percentage score (0-100) based on filled enrichment fields
func (e *Entry) CalculateCompleteness() int {
	score := 0

	// Base: description + evidence (40 points)
	if e.Description != "" {
		score += 20
	}
	if e.Evidence != "" && e.Evidence != "[missing]" {
		score += 20
	}

	// Enrichment fields (15 points each)
	if e.HoursSaved != nil {
		score += 15
	}
	if e.BusinessMetric != "" {
		score += 15
	}
	if e.StrategicAlign != "" {
		score += 15
	}
	if e.PeerRecognition != "" {
		score += 15
	}

	return score
}

// getBragPath is overridable for testing
var getBragPath = func() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "eddie-brag", "brag.json"), nil
}

// readBragDocument reads and validates the brag document
func readBragDocument() (*BragDocument, error) {
	path, err := getBragPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	var doc BragDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		log.Fatalf("FATAL: brag.json is corrupted at %s\nError: %v\nPreserving file for manual recovery. Do NOT run init.", path, err)
	}

	// Self-healing NextID check
	maxID := 0
	for _, entry := range doc.Entries {
		if entry.ID > maxID {
			maxID = entry.ID
		}
	}
	if doc.NextID <= maxID {
		doc.NextID = maxID + 1
	}

	return &doc, nil
}

// writeBragDocument writes the document and auto-commits to git
func writeBragDocument(doc *BragDocument, commitMsg string) error {
	path, err := getBragPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	// Auto-commit and push
	gitDir := filepath.Dir(path)
	if err := gitCommitAndPush(gitDir, commitMsg); err != nil {
		fmt.Printf("Warning: git commit/push failed: %v\n", err)
		fmt.Println("Changes saved locally but not backed up to remote.")
	}

	return nil
}

// gitCommitAndPush commits and pushes changes
func gitCommitAndPush(dir, message string) error {
	// Add all changes
	cmd := exec.Command("git", "-C", dir, "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	cmd = exec.Command("git", "-C", dir, "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		// It's okay if there's nothing to commit
		if !strings.Contains(err.Error(), "nothing to commit") {
			return fmt.Errorf("git commit failed: %w", err)
		}
	}

	// Push
	cmd = exec.Command("git", "-C", dir, "push")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}

// InitBragDocument initializes the brag system
func InitBragDocument(roleTitle string, roleStartDate time.Time) error {
	path, err := getBragPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)

	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("brag.json already exists at %s", path)
	}

	// Initialize git repo
	cmd := exec.Command("git", "-C", dir, "init")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Create private GitHub repo
	cmd = exec.Command("gh", "repo", "create", "edwelker/brag-data", "--private", "--confirm")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: failed to create remote repo: %v\n", err)
		fmt.Println("You can create it manually and add the remote later.")
	}

	// Add remote
	cmd = exec.Command("git", "-C", dir, "remote", "add", "origin", "https://github.com/edwelker/brag-data.git")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: failed to add remote: %v\n", err)
	}

	// Create initial document
	doc := &BragDocument{
		RoleTitle:     roleTitle,
		RoleStartDate: roleStartDate,
		NextID:        1,
		Entries:       []Entry{},
	}

	if err := writeBragDocument(doc, "init: create brag document"); err != nil {
		return err
	}

	fmt.Printf("Initialized brag document at %s\n", path)
	fmt.Println("Backed up to edwelker/brag-data (private)")
	return nil
}

// AddEntry adds a new entry
func AddEntry(bucket, description, evidence, status string, startDate, endDate time.Time) (int, error) {
	doc, err := readBragDocument()
	if err != nil {
		return 0, err
	}

	// Default to "Completed" if not specified (for backward compatibility)
	if status == "" {
		status = "Completed"
	}

	entry := Entry{
		ID:          doc.NextID,
		StartDate:   startDate,
		EndDate:     endDate,
		Bucket:      bucket,
		Description: description,
		Evidence:    evidence,
		Status:      status,
	}

	doc.Entries = append(doc.Entries, entry)
	doc.NextID++

	commitMsg := fmt.Sprintf("add: %s entry #%d", bucket, entry.ID)
	return entry.ID, writeBragDocument(doc, commitMsg)
}

// UpdateEntry updates basic fields of an entry (does not touch enrichment)
func UpdateEntry(id int, bucket, description, evidence, status string, startDate, endDate time.Time) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	found := false
	for i := range doc.Entries {
		if doc.Entries[i].ID == id {
			if bucket != "" {
				doc.Entries[i].Bucket = bucket
			}
			if description != "" {
				doc.Entries[i].Description = description
			}
			if evidence != "" {
				doc.Entries[i].Evidence = evidence
			}
			if status != "" {
				doc.Entries[i].Status = status
			}
			if !startDate.IsZero() {
				doc.Entries[i].StartDate = startDate
			}
			if !endDate.IsZero() {
				doc.Entries[i].EndDate = endDate
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entry #%d not found", id)
	}

	commitMsg := fmt.Sprintf("update: entry #%d", id)
	return writeBragDocument(doc, commitMsg)
}

// RemoveEntry removes an entry by ID
func RemoveEntry(id int) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	found := false
	for i, entry := range doc.Entries {
		if entry.ID == id {
			doc.Entries = append(doc.Entries[:i], doc.Entries[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entry #%d not found", id)
	}

	commitMsg := fmt.Sprintf("remove: entry #%d", id)
	return writeBragDocument(doc, commitMsg)
}

// ClearEntries removes all entries
func ClearEntries() error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	doc.Entries = []Entry{}
	return writeBragDocument(doc, "clear: removed all entries")
}

// EnrichEntry updates enrichment fields for an entry
// Hours saved and calculation are only updated if explicitly provided (non-nil/non-empty)
// Other fields (evidence, metrics, alignment, recognition) are always updated
func EnrichEntry(id int, evidence string, hoursSaved *float64, hoursSavedCalc, businessMetric, strategicAlign, peerRecognition string) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	found := false
	for i := range doc.Entries {
		if doc.Entries[i].ID == id {
			now := time.Now()
			if evidence != "" {
				doc.Entries[i].Evidence = evidence
			}
			// Only update hours/calc if explicitly provided
			if hoursSaved != nil {
				doc.Entries[i].HoursSaved = hoursSaved
			}
			if hoursSavedCalc != "" {
				doc.Entries[i].HoursSavedCalculation = hoursSavedCalc
			}
			// Always update other enrichment fields
			doc.Entries[i].BusinessMetric = businessMetric
			doc.Entries[i].StrategicAlign = strategicAlign
			doc.Entries[i].PeerRecognition = peerRecognition
			doc.Entries[i].EnrichedAt = &now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entry #%d not found", id)
	}

	commitMsg := fmt.Sprintf("enrich: entry #%d", id)
	return writeBragDocument(doc, commitMsg)
}

// UpdateRoleStartDate updates the role start date
func UpdateRoleStartDate(newDate time.Time) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	doc.RoleStartDate = newDate
	return writeBragDocument(doc, "config: update role start date")
}
