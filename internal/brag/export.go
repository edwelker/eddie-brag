package brag

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

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

	path, err := getBragPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
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

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"ID", "StartDate", "EndDate", "Bucket", "Status", "Description", "Evidence", "HoursSaved", "BusinessMetric", "StrategicAlignment", "PeerRecognition"}); err != nil {
		return err
	}

	for _, entry := range entries {
		hours := ""
		if entry.HoursSaved != nil {
			hours = fmt.Sprintf("%.1f", *entry.HoursSaved)
		}

		if err := w.Write([]string{
			fmt.Sprintf("%d", entry.ID),
			entry.StartDate.Format("2006-01-02"),
			entry.EndDate.Format("2006-01-02"),
			entry.Bucket,
			entry.Status,
			entry.Description,
			entry.Evidence,
			hours,
			entry.BusinessMetric,
			entry.StrategicAlign,
			entry.PeerRecognition,
		}); err != nil {
			return err
		}
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
			fmt.Fprintf(f, "#%d [%s] [%s to %s]\n",
				entry.ID,
				entry.Status,
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
