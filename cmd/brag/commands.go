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

func handleInit() {
	var roleStartDateStr string
	prompt := &survey.Input{
		Message: "Role start date (YYYY-MM-DD):",
	}
	if err := survey.AskOne(prompt, &roleStartDateStr, survey.WithValidator(survey.Required)); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	roleStartDate, err := time.ParseInLocation("2006-01-02", roleStartDateStr, time.Local)
	if err != nil {
		fmt.Printf("Invalid date format: %v\n", err)
		os.Exit(1)
	}

	if err := brag.InitBragDocument(roleStartDate); err != nil {
		fmt.Printf("Error initializing: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd() {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	bucket := fs.String("b", "", "Work context bucket")
	description := fs.String("d", "", "Description")
	evidence := fs.String("e", "", "Evidence URL")
	startDateStr := fs.String("start", "", "Start date (YYYY-MM-DD)")
	endDateStr := fs.String("end", "", "End date (YYYY-MM-DD)")
	weekNum := fs.Int("week", 0, "Week number (relative to role start)")
	monthNum := fs.Int("month", 0, "Month number (relative to role start)")

	fs.Parse(os.Args[2:])

	var err error
	var startDate, endDate time.Time

	// Resolve dates first (need role start date from document)
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

	// Interactive prompts if flags not provided
	if *bucket == "" {
		*bucket = promptBucket()
	}

	if *description == "" {
		*description = promptDescription()
	}

	if *evidence == "" {
		*evidence = promptEvidenceWithValidation()
	}

	if err := brag.AddEntry(*bucket, *description, *evidence, startDate, endDate); err != nil {
		fmt.Printf("Error adding entry: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Entry added successfully!")
}

func handleEnrich() {
	fs := flag.NewFlagSet("enrich", flag.ExitOnError)
	rangeStr := fs.String("range", "", "Time range (e.g., 30d)")
	id := fs.Int("id", 0, "Specific entry ID")
	pending := fs.Bool("pending", false, "List pending entries without prompting")

	fs.Parse(os.Args[2:])

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

	// Interactive enrichment
	for _, entry := range entries {
		fmt.Printf("\n#%d [%s to %s] %s\n",
			entry.ID,
			entry.StartDate.Format("2006-01-02"),
			entry.EndDate.Format("2006-01-02"),
			entry.Bucket)
		fmt.Printf("%s\n", entry.Description)

		var proceed bool
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

		// Collect enrichment data
		hoursSaved := promptHoursSaved()
		businessMetric := promptBusinessMetric()
		strategicAlign := promptStrategicAlign()
		peerRecognition := promptPeerRecognition()

		if err := brag.EnrichEntry(entry.ID, hoursSaved, businessMetric, strategicAlign, peerRecognition); err != nil {
			fmt.Printf("Error enriching entry: %v\n", err)
			continue
		}

		fmt.Println("Entry enriched successfully!")
	}
}

func handleList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	rangeStr := fs.String("range", "", "Time range (e.g., 30d)")
	week := fs.Int("week", 0, "Week number")
	month := fs.Int("month", 0, "Month number")
	all := fs.Bool("all", false, "Show all entries")

	fs.Parse(os.Args[2:])

	if err := brag.ListEntries(*rangeStr, *week, *month, *all); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleReport() {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	rangeStr := fs.String("range", "", "Time range (e.g., 90d)")
	week := fs.Int("week", 0, "Week number")
	month := fs.Int("month", 0, "Month number")
	year := fs.Int("year", 0, "Year number")

	fs.Parse(os.Args[2:])

	if err := brag.ReportEntries(*rangeStr, *week, *month, *year); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
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

	fs.Parse(os.Args[2:])

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
