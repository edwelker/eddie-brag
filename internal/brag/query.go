package brag

import (
	"fmt"
	"strings"
	"time"
)

// NoColor disables color output when true
var NoColor = false

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

	// Calculate total hours saved from filtered entries
	var totalHours float64
	for _, entry := range filtered {
		if entry.HoursSaved != nil {
			totalHours += *entry.HoursSaved
		}
	}
	totalBusinessDays := totalHours / 8.0

	// Show role and tenure with color
	tenure := getTenure(doc.RoleStartDate)
	if NoColor {
		fmt.Printf("Role: %s\n", doc.RoleTitle)
		fmt.Printf("Tenure: %s\n", tenure)
		if totalHours > 0 {
			fmt.Printf("Total Hours Saved: %.1f (%.2f business days)\n", totalHours, totalBusinessDays)
		}
	} else {
		fmt.Printf("\033[1;36mRole:\033[0m %s\n", doc.RoleTitle)
		fmt.Printf("\033[1;36mTenure:\033[0m %s\n", tenure)
		if totalHours > 0 {
			fmt.Printf("\033[1;36mTotal Hours Saved:\033[0m %.1f (%.2f business days)\n", totalHours, totalBusinessDays)
		}
	}
	fmt.Println()

	// Group by bucket
	grouped := groupByBucket(filtered)
	bucketOrder := []string{"Delivery", "Architecture", "Process", "Leadership"}

	for _, bucket := range bucketOrder {
		entries, ok := grouped[bucket]
		if !ok || len(entries) == 0 {
			continue
		}

		if NoColor {
			fmt.Printf("## %s\n", bucket)
		} else {
			fmt.Printf("\033[1;35m## %s\033[0m\n", bucket)
		}

		for _, entry := range entries {
			// Skip test/draft entries
			if strings.Contains(strings.ToLower(entry.Status), "test") || strings.Contains(strings.ToLower(entry.Status), "draft") {
				continue
			}

			completeness := entry.CalculateCompleteness()
			enrichMarker := ""
			if entry.Evidence == "" || entry.EnrichedAt == nil {
				if NoColor {
					enrichMarker = " [needs enrichment]"
				} else {
					enrichMarker = " \033[33m[needs enrichment]\033[0m"
				}
			}

			// Show completeness score with color
			completenessIndicator := ""
			if NoColor {
				if completeness < 100 {
					if completeness < 60 {
						completenessIndicator = fmt.Sprintf(" [%d%% ⚠️]", completeness)
					} else if completeness < 90 {
						completenessIndicator = fmt.Sprintf(" [%d%%]", completeness)
					} else {
						completenessIndicator = fmt.Sprintf(" [%d%% ✓]", completeness)
					}
				} else {
					completenessIndicator = " [100% ✓]"
				}
			} else {
				if completeness < 100 {
					if completeness < 60 {
						completenessIndicator = fmt.Sprintf(" \033[31m[%d%% ⚠️]\033[0m", completeness)
					} else if completeness < 90 {
						completenessIndicator = fmt.Sprintf(" \033[33m[%d%%]\033[0m", completeness)
					} else {
						completenessIndicator = fmt.Sprintf(" \033[32m[%d%% ✓]\033[0m", completeness)
					}
				} else {
					completenessIndicator = " \033[32m[100% ✓]\033[0m"
				}
			}

			if NoColor {
				fmt.Printf("#%d [%s] [%s to %s]%s%s\n", entry.ID,
					entry.Status,
					entry.StartDate.Format("2006-01-02"),
					entry.EndDate.Format("2006-01-02"),
					completenessIndicator,
					enrichMarker)
			} else {
				fmt.Printf("\033[1m#%d\033[0m \033[36m[%s]\033[0m [%s to %s]%s%s\n", entry.ID,
					entry.Status,
					entry.StartDate.Format("2006-01-02"),
					entry.EndDate.Format("2006-01-02"),
					completenessIndicator,
					enrichMarker)
			}

			fmt.Printf("  %s\n", entry.Description)
			if NoColor {
				if entry.Evidence != "" {
					fmt.Printf("  Evidence: %s\n", entry.Evidence)
				} else {
					fmt.Printf("  Evidence: [missing]\n")
				}

				if entry.HoursSaved != nil {
					businessDays := *entry.HoursSaved / 8.0
					fmt.Printf("  Hours Saved: %.1f (%.2f business days)\n", *entry.HoursSaved, businessDays)
					if entry.HoursSavedCalculation != "" {
						fmt.Printf("    Calculation: %s\n", entry.HoursSavedCalculation)
					}
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
			} else {
				if entry.Evidence != "" {
					fmt.Printf("  \033[90mEvidence:\033[0m %s\n", entry.Evidence)
				} else {
					fmt.Printf("  \033[90mEvidence:\033[0m \033[31m[missing]\033[0m\n")
				}

				if entry.HoursSaved != nil {
					businessDays := *entry.HoursSaved / 8.0
					fmt.Printf("  \033[90mHours Saved:\033[0m %.1f (%.2f business days)\n", *entry.HoursSaved, businessDays)
					if entry.HoursSavedCalculation != "" {
						fmt.Printf("    \033[90mCalculation:\033[0m %s\n", entry.HoursSavedCalculation)
					}
				}
				if entry.BusinessMetric != "" {
					fmt.Printf("  \033[90mBusiness Metric:\033[0m %s\n", entry.BusinessMetric)
				}
				if entry.StrategicAlign != "" {
					fmt.Printf("  \033[90mStrategic Alignment:\033[0m %s\n", entry.StrategicAlign)
				}
				if entry.PeerRecognition != "" {
					fmt.Printf("  \033[90mPeer Recognition:\033[0m %s\n", entry.PeerRecognition)
				}
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

// GetIncompleteEntries returns all entries with completeness < 100%
func GetIncompleteEntries() ([]Entry, error) {
	doc, err := readBragDocument()
	if err != nil {
		return nil, err
	}

	var incomplete []Entry
	for _, entry := range doc.Entries {
		if entry.CalculateCompleteness() < 100 {
			incomplete = append(incomplete, entry)
		}
	}

	return incomplete, nil
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

	// When targeting specific ID, return it regardless of enrichment status
	// This allows re-enriching already-enriched entries
	if id > 0 {
		return candidates, nil
	}

	// For time-based queries, filter to unenriched only
	var unenriched []Entry
	for _, entry := range candidates {
		if entry.Evidence == "" || entry.EnrichedAt == nil {
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

	fmt.Printf("# Report: %s (%s)\n\n", doc.RoleTitle, periodLabel)

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

		bucketDays := bucketHours / 8.0
		fmt.Printf("## %s (%d entries, %.1f hours / %.2f days saved)\n", bucket, len(entries), bucketHours, bucketDays)
		for _, entry := range entries {
			fmt.Printf("- [%s] %s\n", entry.Status, entry.Description)
		}
		fmt.Println()

		totalEntries += len(entries)
		totalHours += bucketHours
	}

	totalDays := totalHours / 8.0
	fmt.Printf("**Total: %d entries, %.1f hours / %.2f business days saved**\n", totalEntries, totalHours, totalDays)

	return nil
}
