# Architecture Reference

Current implementation decisions (for context when working on the codebase).

## Data Model

```go
type Entry struct {
    ID          int       `json:"id"`
    StartDate   time.Time `json:"start_date"`
    EndDate     time.Time `json:"end_date"`
    Bucket      string    `json:"bucket"`
    Description string    `json:"description"`
    Evidence    string    `json:"evidence"`

    HoursSaved      *float64   `json:"hours_saved,omitempty"`
    BusinessMetric  string     `json:"business_metric,omitempty"`
    StrategicAlign  string     `json:"strategic_alignment,omitempty"`
    PeerRecognition string     `json:"peer_recognition,omitempty"`
    EnrichedAt      *time.Time `json:"enriched_at,omitempty"`
}

type BragDocument struct {
    RoleStartDate time.Time `json:"role_start_date"`
    NextID        int       `json:"next_id"`
    Entries       []Entry   `json:"entries"`
}
```

## Key Design Decisions

- ID is permanent, monotonically increasing, never re-indexed. `NextID` tracks next available.
- Self-healing NextID: on load, compute `max(existing IDs) + 1`. Correct silently if stale.
- Entries span time (start_date + end_date). Default: both = today.
- Allowed buckets: `Delivery`, `Architecture`, `Process`, `Leadership`.
- Enrichment fields use pointers/omitempty to distinguish "not set" from "empty".
- `EnrichedAt` timestamp marks entry as having been through enrich flow (even if fields skipped).
- All date operations use `time.Local` consistently (no UTC drift on week/month boundaries).

## Storage

- Path: `~/.config/eddie-brag/brag.json` (via `os.UserConfigDir()`)
- Single global file regardless of working directory.
- `brag init` creates directory and file if missing.

## Auto-Backup to Git

- `~/.config/eddie-brag/` is a git repo backed by private remote (`edwelker/brag-data`).
- Every write operation: auto-commit + push. Descriptive message (e.g., `add: Delivery entry #42`).
- Push failure (no network): commit locally, warn user. Next write retries.

## Error Handling

- JSON unmarshal failure: `log.Fatalf` with corruption message including file path.
- Never overwrite or reset corrupted file — halt and preserve for manual recovery.

## Date Input

- `--start 2024-01-08 --end 2024-01-12` (explicit ISO)
- `--week 1` (relative to role start date)
- `--month 3` (relative to role start date)
- No flags: both default to today.

## URL Validation

- HTTP HEAD, 5s timeout.
- 200-399: valid (silent). 401/403: "valid but protected" (silent).
- 404, 500+, DNS failure, timeout: warn, ask confirmation.

## Hours Saved Parsing

- Accept: `1.5` (hours), `90m`/`90min` (->1.5h), `2h` (->2.0h), `0` (no savings).
- Reject negative. Store as `*float64`.
