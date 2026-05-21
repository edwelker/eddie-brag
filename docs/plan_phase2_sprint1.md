# Phase 2, Sprint 1: Completeness and Guidance

High impact, low effort. Focus: make it hard to write bad brags.

## 1. Contextual Enrichment Prompts

Template-based prompts per bucket type. Show examples inline during enrich.

Modify `cmd/brag/prompts.go`:
- `promptBusinessMetric(bucket string)` — bucket-specific examples
- `promptStrategicAlign()` — org goal examples
- `promptPeerRecognition()` — Slack link format example

Example output:
```
Business Metric (Process bucket):
  Examples:
    - "Reduced deploy time from 18min to 9min"
    - "Enabled 2x more releases per week"
    - "Saved $50K in compute costs"
    - "Reduced incident count by 30%"

Your metric: _____
```

## 2. Bucket-Specific Required Fields

Conditional validation based on bucket:
- Process bucket: business_metric required (non-empty)
- Leadership bucket: peer_recognition required
- Warn (don't block) if missing, show completeness impact

## 3. Entry Completeness Scoring

Add `calculateCompleteness(entry Entry) int`:
```
Base: 40 points (description + evidence)
+15 points: hours_saved present
+15 points: business_metric present
+15 points: strategic_alignment present
+15 points: peer_recognition present
Total: 100 points
```

Display in `brag list`: `#4 [85% complete]`

## 4. Review Command (`brag review`)

Surfaces incomplete entries (completeness < 100%). For each, prompt to enrich missing fields.

```
Entry #4: 70% complete
  OK: Description, Evidence, Hours saved
  Missing: Business metric, Peer recognition

Add business metric now? [y/N]
```

## Verification

- `brag enrich --id 4` shows example-rich prompts
- Process entry without business_metric warns about completeness
- `brag list` shows `[85% complete]` markers
- `brag review` surfaces incomplete entries and prompts for missing fields
