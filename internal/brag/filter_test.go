package brag

import (
	"testing"
	"time"
)

func TestFilterByDateRange_Overlaps(t *testing.T) {
	// Critical test: entries that span the range boundary should be included
	tests := []struct {
		name       string
		entries    []Entry
		rangeStart time.Time
		rangeEnd   time.Time
		wantIDs    []int
	}{
		{
			name: "entry starts before range, ends within",
			entries: []Entry{
				{ID: 1, StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local)},
			},
			rangeStart: time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local),
			rangeEnd:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
			wantIDs:    []int{1}, // Should be included - overlaps range
		},
		{
			name: "entry starts within range, ends after",
			entries: []Entry{
				{ID: 2, StartDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local)},
			},
			rangeStart: time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local),
			rangeEnd:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
			wantIDs:    []int{2}, // Should be included
		},
		{
			name: "entry spans entire range",
			entries: []Entry{
				{ID: 3, StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 30, 0, 0, 0, 0, time.Local)},
			},
			rangeStart: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local),
			rangeEnd:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local),
			wantIDs:    []int{3}, // Should be included
		},
		{
			name: "entry completely before range",
			entries: []Entry{
				{ID: 4, StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local)},
			},
			rangeStart: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local),
			rangeEnd:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local),
			wantIDs:    []int{}, // Should be excluded
		},
		{
			name: "entry completely after range",
			entries: []Entry{
				{ID: 5, StartDate: time.Date(2024, 1, 25, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 30, 0, 0, 0, 0, time.Local)},
			},
			rangeStart: time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local),
			rangeEnd:   time.Date(2024, 1, 20, 0, 0, 0, 0, time.Local),
			wantIDs:    []int{}, // Should be excluded
		},
		{
			name: "multi-week entry reported in week 1 and week 2",
			entries: []Entry{
				{ID: 6, StartDate: time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local), EndDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local)},
			},
			rangeStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			rangeEnd:   time.Date(2024, 1, 8, 0, 0, 0, 0, time.Local),
			wantIDs:    []int{6}, // Should appear in week 1 report even though it extends to week 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByDateRange(tt.entries, tt.rangeStart, tt.rangeEnd)

			if len(got) != len(tt.wantIDs) {
				t.Errorf("filterByDateRange() returned %d entries, want %d", len(got), len(tt.wantIDs))
			}

			for i, wantID := range tt.wantIDs {
				if i >= len(got) || got[i].ID != wantID {
					t.Errorf("filterByDateRange() entry[%d].ID = %d, want %d", i, got[i].ID, wantID)
				}
			}
		})
	}
}

func TestFilterAfter_BoundaryConditions(t *testing.T) {
	cutoff := time.Date(2024, 1, 10, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name    string
		entries []Entry
		wantIDs []int
	}{
		{
			name: "entry exactly at cutoff",
			entries: []Entry{
				{ID: 1, StartDate: cutoff},
			},
			wantIDs: []int{1}, // Should be included (Equal)
		},
		{
			name: "entry one day before cutoff",
			entries: []Entry{
				{ID: 2, StartDate: cutoff.AddDate(0, 0, -1)},
			},
			wantIDs: []int{}, // Should be excluded
		},
		{
			name: "entry one second after cutoff",
			entries: []Entry{
				{ID: 3, StartDate: cutoff.Add(1 * time.Second)},
			},
			wantIDs: []int{3}, // Should be included
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterAfter(tt.entries, cutoff)

			if len(got) != len(tt.wantIDs) {
				t.Errorf("filterAfter() returned %d entries, want %d", len(got), len(tt.wantIDs))
			}
		})
	}
}
