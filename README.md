# eddie-brag

Staff-level professional accomplishment tracker for performance reviews and career progression.

## What This Is

A CLI tool for tracking work accomplishments with two workflows:
- **Daily capture** (`brag add`) — frictionless, just bucket/description/evidence
- **Weekly synthesis** (`brag enrich`) — add hours saved, business impact, strategic alignment

**Key Features:**
- Color-coded output for quick scanning (green = complete, yellow = needs work, red = critical gaps)
- Smart completeness scoring (focuses on impact metrics, not fluff)
- Auto-sync to private GitHub repo — never lose your career history
- Built for staff+ engineers who need to demonstrate impact across delivery, architecture, process, and leadership

## Installation

**Recommended: Install directly**
```bash
go install github.com/edwelker/eddie-brag/cmd/brag@latest
```
This installs `brag` to `$GOPATH/bin` (usually `~/go/bin`). Make sure it's in your PATH.

**Or: Clone and install from source**
```bash
git clone https://github.com/edwelker/eddie-brag.git
cd eddie-brag
go install ./cmd/brag
```

## Quick Start

```bash
# One-time setup
brag init
# Enter your role title (e.g., Staff Software Engineer)
# Enter your role start date (YYYY-MM-DD)

# Daily: log an accomplishment
brag add

# Weekly: enrich entries with metrics
brag enrich

# View last 7 days
brag list

# Generate month 3 report for review prep
brag report --month 3
```

## Daily Workflow

```bash
# Interactive (default)
brag add

# Flag-based (for scripting)
brag add -b Delivery -d "Shipped feature X" -e "https://jira.example.com/PROJ-123"

# Backfill past work
brag add --week 1 -b Process -d "Fixed CI bottleneck" -e "https://..."
```

## Weekly Enrichment

```bash
# Find unenriched entries from last 7 days
brag enrich

# Check what needs attention
brag enrich --pending

# Enrich last 30 days
brag enrich --range 30d

# Enrich specific entry
brag enrich --id 42
```

Enrichment prompts:
- Hours Saved (accepts `1.5`, `90m`, `2h`)
- Business Metric (e.g., "Reduced P95 latency by 200ms")
- Strategic Alignment (e.g., "Q2 OKR: Improve developer velocity")
- Peer Recognition (e.g., "Slack kudos from @manager")

## Viewing & Reporting

```bash
# Last 7 days (default, color-coded output)
brag list

# Specific time periods
brag list --month 3
brag list --week 12
brag list --range 90d
brag list --all

# Disable colors (for scripts/piping)
brag list --all --no-color

# Generate review-ready report
brag report --month 3  # Grouped by bucket with totals
brag report --year 1   # Full first year summary

# Export to file
brag export --format csv --month 3
brag export --format json --all
```

**Understanding the output:**
- **[100% ✓]** Green checkmark = review-ready (all key fields complete)
- **[90% ✓]** Green = nearly complete, ready for review
- **[75%]** Yellow = needs more enrichment
- **[40% ⚠️]** Red warning = missing critical impact data
- **[needs enrichment]** = hasn't been enriched yet

## Data Storage & Backup

- Data lives at `~/.config/eddie-brag/brag.json`
- Every write auto-commits and pushes to `edwelker/brag-data` (private repo)
- Full git history means you can recover from corruption or accidental edits
- If offline: commits locally, retries push on next write
- **Role changes:** When promoted or changing companies, archive the current file (`mv brag.json brag-staff-engineer-2024.json`) and run `brag init` again to start fresh with your new role title

## Zsh Aliases

Add to `~/.zshrc`:

```bash
alias bradd='brag add'
alias bragf='brag enrich'
alias bragl='brag list'
alias bragr='brag report'

# Weekly export to Markdown for review doc
function brag-weekly() {
  brag export --format txt --range 7d
  cat ~/.config/eddie-brag/brag.txt | pbcopy
  echo "Last 7 days copied to clipboard"
}
```

## Help System

```bash
brag help           # Overview
brag help add       # Detailed add usage
brag help enrich    # Detailed enrich usage
brag help list      # List flags
brag help report    # Report options
brag help export    # Export formats
```

## Architecture (for Go learners)

```
cmd/brag/
  main.go       - CLI dispatcher
  commands.go   - Command handlers (add, list, enrich, export, report)
  prompts.go    - Interactive survey logic
  help.go       - Help text
  validation.go - Input validation

internal/brag/
  brag.go       - Core domain logic, storage, CRUD operations
  query.go      - List/report formatting with color output
  export.go     - JSON/CSV/TXT export
  date.go       - Date math, tenure calculation, business day counting
  filter.go     - Time-based filtering (week/month/range)
  validation.go - URL validation, hours parsing
  *_test.go     - 88% test coverage

.github/workflows/
  test.yml      - CI (runs on macOS + Ubuntu)
```

**Key design decisions:**
- `time.Local` everywhere (no UTC drift on week/month boundaries)
- Self-healing `NextID` (protects against manual JSON edits)
- `EnrichedAt` timestamp (prevents infinite re-prompt loop)
- URL validation treats 401/403 as valid (internal tools are auth-protected)
- Auto git commit+push on every write (zero-effort backup)
- Completeness scoring emphasizes impact over fluff:
  - Description (20%), Evidence (20%), Hours/Calculation (20%), Business Metric (20%), Strategic Alignment (20%)
  - Peer recognition is bonus context, not scored (prevents gaming the metric)
  - Zero hours + explanation = complete (some work prevents future waste)

## License

MIT
