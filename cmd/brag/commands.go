package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/edwelker/eddie-brag/internal/brag"
)

const (
	bucketProcess    = "Process"
	bucketLeadership = "Leadership"
)

func handleInit() {
	var roleTitle string
	titlePrompt := &survey.Input{
		Message: "Role title:",
	}
	if err := survey.AskOne(titlePrompt, &roleTitle, survey.WithValidator(survey.Required)); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var roleStartDateStr string
	datePrompt := &survey.Input{
		Message: "Role start date (YYYY-MM-DD):",
	}
	if err := survey.AskOne(datePrompt, &roleStartDateStr, survey.WithValidator(survey.Required)); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	roleStartDate, err := time.ParseInLocation("2006-01-02", roleStartDateStr, time.Local)
	if err != nil {
		fmt.Printf("Invalid date format: %v\n", err)
		os.Exit(1)
	}

	if err := brag.InitBragDocument(roleTitle, roleStartDate); err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd() {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	bucket := fs.String("b", "", "Work context bucket")
	description := fs.String("d", "", "Description")
	evidence := fs.String("e", "", "Evidence URL")
	status := fs.String("status", "", "Status (Completed, In Progress, Proposed, Abandoned)")
	startDateStr := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDateStr := fs.String("end", "", "End date (YYYY-MM-DD)")
	weekNum := fs.Int("week", 0, "Week number (relative to role start)")
	monthNum := fs.Int("month", 0, "Month number (relative to role start)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	var err error
	var startDate, endDate time.Time

	// Resolve dates first (need role start date from document)
	doc, err := readDocForDates()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check if any flags provided for non-interactive mode
	hasDateFlags := *weekNum > 0 || *monthNum > 0 || *startDateStr != "" || *endDateStr != ""
	hasContentFlags := *bucket != "" || *description != ""
	interactiveMode := !hasDateFlags && !hasContentFlags

	if hasDateFlags {
		// Use flags for dates
		startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, *weekNum, *monthNum, *startDateStr, *endDateStr)
		if err != nil {
			fmt.Printf("Error resolving dates: %v\n", err)
			os.Exit(1)
		}
	} else if interactiveMode {
		// Interactive date prompt
		dateChoice := promptDateOption()
		switch dateChoice {
		case "Today":
			startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, 0, 0, "", "")
			if err != nil {
				fmt.Printf("Error resolving today's date: %v\n", err)
				os.Exit(1)
			}
		case "Specific date":
			*startDateStr = promptSpecificDate("Start date")
			*endDateStr = promptSpecificDate("End date")
			startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, 0, 0, *startDateStr, *endDateStr)
			if err != nil {
				fmt.Printf("Error parsing dates: %v\n", err)
				os.Exit(1)
			}
		case "Week number":
			*weekNum = promptWeekNumber()
			startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, *weekNum, 0, "", "")
			if err != nil {
				fmt.Printf("Error resolving week dates: %v\n", err)
				os.Exit(1)
			}
		case "Month number":
			*monthNum = promptMonthNumber()
			startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, 0, *monthNum, "", "")
			if err != nil {
				fmt.Printf("Error resolving month dates: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Flags provided but no dates - default to today
		startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, *weekNum, *monthNum, *startDateStr, *endDateStr)
		if err != nil {
			fmt.Printf("Error resolving dates: %v\n", err)
			os.Exit(1)
		}
	}

	// Interactive prompts for content if flags not provided
	if *bucket == "" {
		*bucket = promptBucket()
	}

	if *description == "" {
		*description = promptDescription()
	}

	// Evidence validation - warn if missing
	if *evidence == "" && interactiveMode {
		*evidence = promptEvidenceWithOptions()
	}

	// Status prompt in interactive mode
	if *status == "" && interactiveMode {
		*status = promptStatus()
	}

	newID, err := brag.AddEntry(*bucket, *description, *evidence, *status, startDate, endDate)
	if err != nil {
		fmt.Printf("Error adding entry: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Entry added successfully!")

	// Ask if user wants to enrich now
	if interactiveMode && promptEnrichNow() {
		hoursSaved, hoursSavedCalc := promptHoursSaved()

		// Bucket-specific prompting for critical fields
		var businessMetric string
		if *bucket == bucketProcess {
			businessMetric = promptRequiredBusinessMetric(*bucket)
		} else {
			businessMetric = promptBusinessMetric(*bucket)
		}

		strategicAlign := promptStrategicAlign()

		var peerRecognition string
		if *bucket == bucketLeadership {
			peerRecognition = promptRequiredPeerRecognition()
		} else {
			peerRecognition = promptPeerRecognition()
		}

		// Validate before saving
		if !validateEnrichment(*bucket, businessMetric, peerRecognition) {
			fmt.Println("Enrichment cancelled.")
			return
		}

		if err := brag.EnrichEntry(newID, *evidence, hoursSaved, hoursSavedCalc, businessMetric, strategicAlign, peerRecognition); err != nil {
			fmt.Printf("Error enriching entry: %v\n", err)
			return
		}

		fmt.Println("Entry enriched successfully!")
	}
}

func handleUpdate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: brag update <id> [options]")
		fmt.Println("Options: -b bucket, -d description, -e evidence, --start date, --end date")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid ID: %s\n", os.Args[2])
		os.Exit(1)
	}

	fs := flag.NewFlagSet("update", flag.ExitOnError)
	bucket := fs.String("b", "", "Work context bucket")
	description := fs.String("d", "", "Description")
	evidence := fs.String("e", "", "Evidence URL")
	status := fs.String("status", "", "Status (Completed, In Progress, Proposed, Abandoned)")
	startDateStr := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDateStr := fs.String("end", "", "End date (YYYY-MM-DD)")
	weekNum := fs.Int("week", 0, "Week number (relative to role start)")
	monthNum := fs.Int("month", 0, "Month number (relative to role start)")

	if err := fs.Parse(os.Args[3:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	var startDate, endDate time.Time

	// Resolve dates if provided
	if *weekNum > 0 || *monthNum > 0 || *startDateStr != "" || *endDateStr != "" {
		doc, err := readDocForDates()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		startDate, endDate, err = brag.ResolveDateFlags(doc.RoleStartDate, *weekNum, *monthNum, *startDateStr, *endDateStr)
		if err != nil {
			fmt.Printf("Error resolving dates: %v\n", err)
			os.Exit(1)
		}
	}

	if err := brag.UpdateEntry(id, *bucket, *description, *evidence, *status, startDate, endDate); err != nil {
		fmt.Printf("Error updating entry: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Entry updated successfully!")
}

func handleEnrich() {
	fs := flag.NewFlagSet("enrich", flag.ExitOnError)
	rangeStr := fs.String("range", "", "Time range (e.g., 30d)")
	id := fs.Int("id", 0, "Specific entry ID")
	pending := fs.Bool("pending", false, "List pending entries without prompting")
	hoursStr := fs.String("hours", "", "Hours saved (e.g., 29.625)")
	calc := fs.String("calc", "", "Calculation notes")
	metric := fs.String("metric", "", "Business metric")
	align := fs.String("align", "", "Strategic alignment")
	recognition := fs.String("recognition", "", "Peer recognition")
	skipPrompt := fs.Bool("yes", false, "Skip confirmation prompt")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	entries, err := brag.GetUnenrichedEntries(*rangeStr, *id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("No unenriched entries found.")
		return
	}

	if *pending {
		// Just list them
		fmt.Printf("Found %d unenriched entries:\n\n", len(entries))
		for _, entry := range entries {
			fmt.Printf("#%d [%s] %s - %s\n",
				entry.ID,
				entry.Bucket,
				entry.StartDate.Format("2006-01-02"),
				entry.Description)
		}
		return
	}

	// Check if flags provide all data for non-interactive mode
	hasFlags := *hoursStr != "" || *calc != "" || *metric != "" || *align != "" || *recognition != ""

	// Interactive enrichment
	for _, entry := range entries {
		fmt.Printf("\n#%d [%s to %s] %s\n",
			entry.ID,
			entry.StartDate.Format("2006-01-02"),
			entry.EndDate.Format("2006-01-02"),
			entry.Bucket)
		fmt.Printf("%s\n", entry.Description)

		// Skip confirmation if --yes flag or if using flags for data
		proceed := *skipPrompt || hasFlags
		if !proceed {
			confirmPrompt := &survey.Confirm{
				Message: "Enrich this entry?",
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &proceed); err != nil {
				continue
			}

			if !proceed {
				continue
			}
		}

		// Prompt for evidence if missing and not provided via flag
		evidence := entry.Evidence
		if evidence == "" {
			evidence = promptEvidenceWithValidation()
		}

		// Collect enrichment data from flags or prompts
		var hoursSaved *float64
		var hoursSavedCalc string
		var businessMetric string
		var strategicAlign string
		var peerRecognition string

		if hasFlags {
			// Use flag values
			if *hoursStr != "" {
				hours, err := brag.ParseHoursInput(*hoursStr)
				if err != nil {
					fmt.Printf("Error parsing hours: %v\n", err)
					continue
				}
				hoursSaved = &hours
			}
			hoursSavedCalc = *calc
			businessMetric = *metric
			strategicAlign = *align
			peerRecognition = *recognition
		} else {
			// Interactive prompts - show existing enrichment data first
			displayExistingEnrichment(entry)

			// Prompt for status update
			newStatus := promptStatusUpdate(entry.Status)
			if newStatus != "" {
				if err := brag.UpdateEntry(entry.ID, "", "", "", newStatus, time.Time{}, time.Time{}); err != nil {
					fmt.Printf("Error updating status: %v\n", err)
				} else {
					fmt.Printf("Status updated to '%s'\n", newStatus)
				}
			}

			// Use context-aware prompts that show existing values
			hoursSaved, hoursSavedCalc = promptHoursSavedWithExisting(entry.HoursSaved, entry.HoursSavedCalculation)

			// Bucket-specific prompting for critical fields
			if entry.Bucket == bucketProcess {
				// For Process bucket, business metric is required
				businessMetric = promptBusinessMetricWithExisting(entry.Bucket, entry.BusinessMetric)
				if businessMetric == "" && entry.BusinessMetric == "" {
					businessMetric = promptRequiredBusinessMetric(entry.Bucket)
				}
			} else {
				businessMetric = promptBusinessMetricWithExisting(entry.Bucket, entry.BusinessMetric)
			}

			strategicAlign = promptStrategicAlignWithExisting(entry.StrategicAlign)

			if entry.Bucket == bucketLeadership {
				// For Leadership bucket, peer recognition is required
				peerRecognition = promptPeerRecognitionWithExisting(entry.PeerRecognition)
				if peerRecognition == "" && entry.PeerRecognition == "" {
					peerRecognition = promptRequiredPeerRecognition()
				}
			} else {
				peerRecognition = promptPeerRecognitionWithExisting(entry.PeerRecognition)
			}

			// Validate before saving - only validate new values if they're being set
			if businessMetric != "" || peerRecognition != "" {
				if !validateEnrichment(entry.Bucket, businessMetric, peerRecognition) {
					fmt.Println("Enrichment cancelled.")
					continue
				}
			}
		}

		if err := brag.EnrichEntry(entry.ID, evidence, hoursSaved, hoursSavedCalc, businessMetric, strategicAlign, peerRecognition); err != nil {
			fmt.Printf("Error enriching entry: %v\n", err)
			continue
		}

		fmt.Println("Entry enriched successfully!")
	}
}

func handleList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	rangeStr := fs.String("range", "", "Time range (e.g., 30d)")
	week := fs.Int("week", -1, "Week number (-1 = no filter, 0 = current week)")
	month := fs.Int("month", -1, "Month number (-1 = no filter, 0 = current month)")
	all := fs.Bool("all", false, "Show all entries")
	noColor := fs.Bool("no-color", false, "Disable color output")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	brag.NoColor = *noColor

	// Handle 0 as "current period"
	if *week == 0 || *month == 0 {
		doc, err := readDocForDates()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if *week == 0 {
			*week = brag.GetCurrentWeek(doc.RoleStartDate)
		}
		if *month == 0 {
			*month = brag.GetCurrentMonth(doc.RoleStartDate)
		}
	}

	// Convert -1 back to 0 for "no filter"
	if *week == -1 {
		*week = 0
	}
	if *month == -1 {
		*month = 0
	}

	if err := brag.ListEntries(*rangeStr, *week, *month, *all); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleReport() {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	rangeStr := fs.String("range", "", "Time range (e.g., 90d)")
	week := fs.Int("week", -1, "Week number (-1 = no filter, 0 = current week)")
	month := fs.Int("month", -1, "Month number (-1 = no filter, 0 = current month)")
	year := fs.Int("year", -1, "Year number (-1 = no filter, 0 = current year)")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Handle 0 as "current period"
	if *week == 0 || *month == 0 || *year == 0 {
		doc, err := readDocForDates()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if *week == 0 {
			*week = brag.GetCurrentWeek(doc.RoleStartDate)
		}
		if *month == 0 {
			*month = brag.GetCurrentMonth(doc.RoleStartDate)
		}
		if *year == 0 {
			*year = brag.GetCurrentYear(doc.RoleStartDate)
		}
	}

	// Convert -1 back to 0 for "no filter"
	if *week == -1 {
		*week = 0
	}
	if *month == -1 {
		*month = 0
	}
	if *year == -1 {
		*year = 0
	}

	if err := brag.ReportEntries(*rangeStr, *week, *month, *year); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleReview() {
	entries, err := brag.GetIncompleteEntries()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("🎉 All entries are 100% complete!")
		return
	}

	fmt.Printf("Found %d incomplete entries:\n\n", len(entries))

	for _, entry := range entries {
		completeness := entry.CalculateCompleteness()

		fmt.Printf("Entry #%d [%s] [%s] - %d%% complete\n", entry.ID, entry.Bucket, entry.Status, completeness)
		fmt.Printf("  %s\n", entry.Description)
		fmt.Println()

		// Show what's present and what's missing
		fmt.Println("  Fields:")
		if entry.Description != "" {
			fmt.Println("    ✓ Description")
		}
		if entry.Evidence != "" && entry.Evidence != "[missing]" {
			fmt.Println("    ✓ Evidence")
		} else {
			fmt.Println("    ✗ Evidence missing")
		}
		if entry.HoursSaved != nil {
			fmt.Println("    ✓ Hours saved")
		} else {
			fmt.Println("    ✗ Hours saved missing")
		}
		if entry.BusinessMetric != "" {
			fmt.Println("    ✓ Business metric")
		} else {
			fmt.Println("    ✗ Business metric missing")
		}
		if entry.StrategicAlign != "" {
			fmt.Println("    ✓ Strategic alignment")
		} else {
			fmt.Println("    ✗ Strategic alignment missing")
		}
		if entry.PeerRecognition != "" {
			fmt.Println("    ✓ Peer recognition")
		} else {
			fmt.Println("    ✗ Peer recognition missing")
		}
		fmt.Println()

		// Ask if they want to enrich now
		var enrichNow bool
		enrichPrompt := &survey.Confirm{
			Message: "Improve this entry now?",
			Default: false,
		}
		if err := survey.AskOne(enrichPrompt, &enrichNow); err != nil {
			continue
		}

		if !enrichNow {
			continue
		}

		// Enrich missing fields only

		// Prompt for status update first
		newStatus := promptStatusUpdate(entry.Status)
		if newStatus != "" {
			if err := brag.UpdateEntry(entry.ID, "", "", "", newStatus, time.Time{}, time.Time{}); err != nil {
				fmt.Printf("Error updating status: %v\n", err)
			} else {
				fmt.Printf("Status updated to '%s'\n", newStatus)
			}
		}

		var evidence string
		if entry.Evidence == "" || entry.Evidence == "[missing]" {
			evidence = promptEvidenceWithValidation()
		}

		var hoursSaved *float64
		var hoursSavedCalc string
		if entry.HoursSaved == nil {
			hoursSaved, hoursSavedCalc = promptHoursSaved()
		}

		var businessMetric string
		if entry.BusinessMetric == "" {
			if entry.Bucket == bucketProcess {
				businessMetric = promptRequiredBusinessMetric(entry.Bucket)
			} else {
				businessMetric = promptBusinessMetric(entry.Bucket)
			}
		}

		var strategicAlign string
		if entry.StrategicAlign == "" {
			strategicAlign = promptStrategicAlign()
		}

		var peerRecognition string
		if entry.PeerRecognition == "" {
			if entry.Bucket == bucketLeadership {
				peerRecognition = promptRequiredPeerRecognition()
			} else {
				peerRecognition = promptPeerRecognition()
			}
		}

		// Validate
		if !validateEnrichment(entry.Bucket, businessMetric, peerRecognition) {
			fmt.Println("Skipped.")
			continue
		}

		if err := brag.EnrichEntry(entry.ID, evidence, hoursSaved, hoursSavedCalc, businessMetric, strategicAlign, peerRecognition); err != nil {
			fmt.Printf("Error enriching entry: %v\n", err)
			continue
		}

		fmt.Println("✓ Entry improved!")
		fmt.Println()
	}

	fmt.Println("Review complete.")
}

func handleRemove() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: brag remove <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid ID: %v\n", err)
		os.Exit(1)
	}

	if err := brag.RemoveEntry(id); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Entry removed successfully!")
}

func handleClear() {
	var confirm bool
	prompt := &survey.Confirm{
		Message: "Are you sure you want to clear all entries?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirm); err != nil || !confirm {
		fmt.Println("Cancelled.")
		return
	}

	if err := brag.ClearEntries(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All entries cleared!")
}

func handleExport() {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	format := fs.String("format", "csv", "Export format (csv, json, txt)")
	rangeStr := fs.String("range", "", "Time range")
	week := fs.Int("week", 0, "Week number")
	month := fs.Int("month", 0, "Month number")
	all := fs.Bool("all", false, "Export all entries")

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if err := brag.ExportEntries(*format, *rangeStr, *week, *month, *all); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: brag config start-date <YYYY-MM-DD>")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "start-date":
		if len(os.Args) < 4 {
			fmt.Println("Usage: brag config start-date <YYYY-MM-DD>")
			os.Exit(1)
		}

		dateStr := os.Args[3]
		newDate, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
		if err != nil {
			fmt.Printf("Invalid date format: %v\n", err)
			os.Exit(1)
		}

		if err := brag.UpdateRoleStartDate(newDate); err != nil {
			fmt.Printf("Error updating start date: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Role start date updated successfully!")

	default:
		fmt.Printf("Unknown config command: %s\n", subcommand)
		os.Exit(1)
	}
}

func handleHelp() {
	if len(os.Args) < 3 {
		printUsage()
		return
	}

	command := os.Args[2]

	switch command {
	case "add":
		printAddHelp()
	case "update":
		printUpdateHelp()
	case "enrich":
		printEnrichHelp()
	case "list":
		printListHelp()
	case "report":
		printReportHelp()
	case "export":
		printExportHelp()
	default:
		printUsage()
	}
}

// Helper to read document for date resolution
func readDocForDates() (*struct{ RoleStartDate time.Time }, error) {
	// Minimal struct just for reading role start date
	type minimalDoc struct {
		RoleStartDate time.Time `json:"role_start_date"`
	}

	path, err := getBragPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read brag.json: %w", err)
	}

	var doc minimalDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse brag.json: %w", err)
	}

	return &struct{ RoleStartDate time.Time }{RoleStartDate: doc.RoleStartDate}, nil
}

func getBragPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "eddie-brag", "brag.json"), nil
}
