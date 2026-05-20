package brag

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Entry represents a single accomplishment entry
type Entry struct {
	ID          int        `json:"id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	Bucket      string     `json:"bucket"`
	Description string     `json:"description"`
	Evidence    string     `json:"evidence"`
	HoursSaved      *float64   `json:"hours_saved,omitempty"`
	BusinessMetric  string     `json:"business_metric,omitempty"`
	StrategicAlign  string     `json:"strategic_alignment,omitempty"`
	PeerRecognition string     `json:"peer_recognition,omitempty"`
	EnrichedAt      *time.Time `json:"enriched_at,omitempty"`
}

// BragDocument represents the entire brag document
type BragDocument struct {
	RoleStartDate time.Time `json:"role_start_date"`
	NextID        int       `json:"next_id"`
	Entries       []Entry   `json:"entries"`
}

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

// getBragPath returns the path to brag.json
func getBragPath() (string, error) {
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
func InitBragDocument(roleStartDate time.Time) error {
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
func AddEntry(bucket, description, evidence string, startDate, endDate time.Time) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	entry := Entry{
		ID:          doc.NextID,
		StartDate:   startDate,
		EndDate:     endDate,
		Bucket:      bucket,
		Description: description,
		Evidence:    evidence,
	}

	doc.Entries = append(doc.Entries, entry)
	doc.NextID++

	commitMsg := fmt.Sprintf("add: %s entry #%d", bucket, entry.ID)
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
func EnrichEntry(id int, hoursSaved *float64, businessMetric, strategicAlign, peerRecognition string) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	found := false
	for i := range doc.Entries {
		if doc.Entries[i].ID == id {
			now := time.Now()
			doc.Entries[i].HoursSaved = hoursSaved
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

// ListEntries lists entries with optional time filtering
func ListEntries(rangeStr string, weekNum, monthNum int, all bool) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	// Filter entries
	var filtered []Entry
	cutoff := time.Now().Add(-7 * 24 * time.Hour) // Default: 7 days

	if all {
		filtered = doc.Entries
	} else if weekNum > 0 {
		start, end := getWeekRange(doc.RoleStartDate, weekNum)
		filtered = filterByDateRange(doc.Entries, start, end)
	} else if monthNum > 0 {
		start, end := getMonthRange(doc.RoleStartDate, monthNum)
		filtered = filterByDateRange(doc.Entries, start, end)
	} else if rangeStr != "" {
		dur, err := parseDuration(rangeStr)
		if err != nil {
			return fmt.Errorf("invalid range: %w", err)
		}
		cutoff = time.Now().Add(-dur)
		filtered = filterAfter(doc.Entries, cutoff)
	} else {
		filtered = filterAfter(doc.Entries, cutoff)
	}

	if len(filtered) == 0 {
		fmt.Println("No entries found.")
		return nil
	}

	// Show tenure
	tenure := getTenure(doc.RoleStartDate)
	fmt.Printf("Tenure: %s\n\n", tenure)

	// Group by bucket
	grouped := groupByBucket(filtered)
	bucketOrder := []string{"Delivery", "Architecture", "Process", "Leadership"}

	for _, bucket := range bucketOrder {
		entries, ok := grouped[bucket]
		if !ok || len(entries) == 0 {
			continue
		}

		fmt.Printf("## %s\n", bucket)
		for _, entry := range entries {
			enrichMarker := ""
			if entry.EnrichedAt == nil {
				enrichMarker = " [needs enrichment]"
			}

			fmt.Printf("#%d [%s to %s]%s\n", entry.ID,
				entry.StartDate.Format("2006-01-02"),
				entry.EndDate.Format("2006-01-02"),
				enrichMarker)
			fmt.Printf("  %s\n", entry.Description)
			fmt.Printf("  Evidence: %s\n", entry.Evidence)

			if entry.HoursSaved != nil {
				fmt.Printf("  Hours Saved: %.1f\n", *entry.HoursSaved)
			}
			if entry.BusinessMetric != "" {
				fmt.Printf("  Business Metric: %s\n", entry.BusinessMetric)
			}
			if entry.StrategicAlign != "" {
				fmt.Printf("  Strategic Alignment: %s\n", entry.StrategicAlign)
			}
			if entry.PeerRecognition != "" {
				fmt.Printf("  Peer Recognition: %s\n", entry.PeerRecognition)
			}
			fmt.Println()
		}
	}

	// Hint
	if !all && weekNum == 0 && monthNum == 0 && rangeStr == "" {
		fmt.Println("Showing last 7 days. Use --range 30d, --month 3, or --all to view older entries.")
	}

	return nil
}

// GetUnenrichedEntries returns entries that need enrichment
func GetUnenrichedEntries(rangeStr string, id int) ([]Entry, error) {
	doc, err := readBragDocument()
	if err != nil {
		return nil, err
	}

	var candidates []Entry

	if id > 0 {
		// Specific ID
		for _, entry := range doc.Entries {
			if entry.ID == id {
				candidates = []Entry{entry}
				break
			}
		}
	} else if rangeStr != "" {
		// Time range
		dur, err := parseDuration(rangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid range: %w", err)
		}
		cutoff := time.Now().Add(-dur)
		candidates = filterAfter(doc.Entries, cutoff)
	} else {
		// Default: last 7 days
		cutoff := time.Now().Add(-7 * 24 * time.Hour)
		candidates = filterAfter(doc.Entries, cutoff)
	}

	// Filter to unenriched
	var unenriched []Entry
	for _, entry := range candidates {
		if entry.EnrichedAt == nil {
			unenriched = append(unenriched, entry)
		}
	}

	return unenriched, nil
}

// ReportEntries generates a summary report
func ReportEntries(rangeStr string, weekNum, monthNum, yearNum int) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	var filtered []Entry
	var periodLabel string

	if weekNum > 0 {
		start, end := getWeekRange(doc.RoleStartDate, weekNum)
		filtered = filterByDateRange(doc.Entries, start, end)
		periodLabel = fmt.Sprintf("Week %d", weekNum)
	} else if monthNum > 0 {
		start, end := getMonthRange(doc.RoleStartDate, monthNum)
		filtered = filterByDateRange(doc.Entries, start, end)
		periodLabel = fmt.Sprintf("Month %d", monthNum)
	} else if yearNum > 0 {
		start, end := getYearRange(doc.RoleStartDate, yearNum)
		filtered = filterByDateRange(doc.Entries, start, end)
		periodLabel = fmt.Sprintf("Year %d", yearNum)
	} else if rangeStr != "" {
		dur, err := parseDuration(rangeStr)
		if err != nil {
			return fmt.Errorf("invalid range: %w", err)
		}
		cutoff := time.Now().Add(-dur)
		filtered = filterAfter(doc.Entries, cutoff)
		periodLabel = fmt.Sprintf("Last %s", rangeStr)
	} else {
		return fmt.Errorf("must specify --week, --month, --year, or --range")
	}

	if len(filtered) == 0 {
		fmt.Printf("No entries found for %s.\n", periodLabel)
		return nil
	}

	fmt.Printf("# Report: %s\n\n", periodLabel)

	grouped := groupByBucket(filtered)
	bucketOrder := []string{"Delivery", "Architecture", "Process", "Leadership"}

	totalEntries := 0
	totalHours := 0.0

	for _, bucket := range bucketOrder {
		entries, ok := grouped[bucket]
		if !ok || len(entries) == 0 {
			continue
		}

		bucketHours := 0.0
		for _, entry := range entries {
			if entry.HoursSaved != nil {
				bucketHours += *entry.HoursSaved
			}
		}

		fmt.Printf("## %s (%d entries, %.1f hours saved)\n", bucket, len(entries), bucketHours)
		for _, entry := range entries {
			fmt.Printf("- %s\n", entry.Description)
		}
		fmt.Println()

		totalEntries += len(entries)
		totalHours += bucketHours
	}

	fmt.Printf("**Total: %d entries, %.1f hours saved**\n", totalEntries, totalHours)

	return nil
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

// Helper functions

func groupByBucket(entries []Entry) map[string][]Entry {
	grouped := make(map[string][]Entry)
	for _, entry := range entries {
		grouped[entry.Bucket] = append(grouped[entry.Bucket], entry)
	}
	return grouped
}

func filterAfter(entries []Entry, cutoff time.Time) []Entry {
	var filtered []Entry
	for _, entry := range entries {
		if entry.StartDate.After(cutoff) || entry.StartDate.Equal(cutoff) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func filterByDateRange(entries []Entry, start, end time.Time) []Entry {
	var filtered []Entry
	for _, entry := range entries {
		// Check if entry overlaps with range
		if entry.StartDate.Before(end) && entry.EndDate.After(start) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
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

func getTenure(roleStart time.Time) string {
	now := time.Now()
	diff := now.Sub(roleStart)

	days := int(diff.Hours() / 24)
	weeks := days / 7
	months := int(now.Sub(roleStart).Hours() / 24 / 30.44) // Average month length

	return fmt.Sprintf("Week %d | Month %d | Day %d since %s",
		weeks, months, days, roleStart.Format("2006-01-02"))
}

func parseDuration(s string) (time.Duration, error) {
	// Parse formats like "30d", "12w", "3m"
	re := regexp.MustCompile(`^(\d+)([dwmy])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format (use 30d, 12w, etc)")
	}

	num, _ := strconv.Atoi(matches[1])
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

	return 0, fmt.Errorf("unknown unit: %s", unit)
}

// ExportEntries exports to file
func ExportEntries(format, rangeStr string, weekNum, monthNum int, all bool) error {
	doc, err := readBragDocument()
	if err != nil {
		return err
	}

	// Filter entries (same logic as list)
	var filtered []Entry
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	if all {
		filtered = doc.Entries
	} else if weekNum > 0 {
		start, end := getWeekRange(doc.RoleStartDate, weekNum)
		filtered = filterByDateRange(doc.Entries, start, end)
	} else if monthNum > 0 {
		start, end := getMonthRange(doc.RoleStartDate, monthNum)
		filtered = filterByDateRange(doc.Entries, start, end)
	} else if rangeStr != "" {
		dur, err := parseDuration(rangeStr)
		if err != nil {
			return fmt.Errorf("invalid range: %w", err)
		}
		cutoff = time.Now().Add(-dur)
		filtered = filterAfter(doc.Entries, cutoff)
	} else {
		filtered = filterAfter(doc.Entries, cutoff)
	}

	// Sort by bucket
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Bucket < filtered[j].Bucket
	})

	path, _ := getBragPath()
	exportDir := filepath.Dir(path)
	exportPath := filepath.Join(exportDir, fmt.Sprintf("brag.%s", format))

	switch format {
	case "json":
		return exportJSON(exportPath, filtered)
	case "csv":
		return exportCSV(exportPath, filtered)
	case "txt":
		return exportTXT(exportPath, filtered)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func exportJSON(path string, entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Exported to %s\n", path)
	return nil
}

func exportCSV(path string, entries []Entry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Header
	fmt.Fprintln(f, "ID,StartDate,EndDate,Bucket,Description,Evidence,HoursSaved,BusinessMetric,StrategicAlignment,PeerRecognition")

	for _, entry := range entries {
		hours := ""
		if entry.HoursSaved != nil {
			hours = fmt.Sprintf("%.1f", *entry.HoursSaved)
		}

		fmt.Fprintf(f, "%d,%s,%s,%s,\"%s\",\"%s\",%s,\"%s\",\"%s\",\"%s\"\n",
			entry.ID,
			entry.StartDate.Format("2006-01-02"),
			entry.EndDate.Format("2006-01-02"),
			entry.Bucket,
			strings.ReplaceAll(entry.Description, "\"", "\"\""),
			strings.ReplaceAll(entry.Evidence, "\"", "\"\""),
			hours,
			strings.ReplaceAll(entry.BusinessMetric, "\"", "\"\""),
			strings.ReplaceAll(entry.StrategicAlign, "\"", "\"\""),
			strings.ReplaceAll(entry.PeerRecognition, "\"", "\"\""),
		)
	}

	fmt.Printf("Exported to %s\n", path)
	return nil
}

func exportTXT(path string, entries []Entry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	grouped := groupByBucket(entries)
	bucketOrder := []string{"Delivery", "Architecture", "Process", "Leadership"}

	for _, bucket := range bucketOrder {
		if len(grouped[bucket]) == 0 {
			continue
		}

		fmt.Fprintf(f, "## %s\n\n", bucket)
		for _, entry := range grouped[bucket] {
			fmt.Fprintf(f, "#%d [%s to %s]\n",
				entry.ID,
				entry.StartDate.Format("2006-01-02"),
				entry.EndDate.Format("2006-01-02"))
			fmt.Fprintf(f, "%s\n", entry.Description)
			fmt.Fprintf(f, "Evidence: %s\n", entry.Evidence)

			if entry.HoursSaved != nil {
				fmt.Fprintf(f, "Hours Saved: %.1f\n", *entry.HoursSaved)
			}
			if entry.BusinessMetric != "" {
				fmt.Fprintf(f, "Business Metric: %s\n", entry.BusinessMetric)
			}
			if entry.StrategicAlign != "" {
				fmt.Fprintf(f, "Strategic Alignment: %s\n", entry.StrategicAlign)
			}
			if entry.PeerRecognition != "" {
				fmt.Fprintf(f, "Peer Recognition: %s\n", entry.PeerRecognition)
			}
			fmt.Fprintln(f)
		}
	}

	fmt.Printf("Exported to %s\n", path)
	return nil
}
