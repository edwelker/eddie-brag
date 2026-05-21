package brag

import "time"

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
