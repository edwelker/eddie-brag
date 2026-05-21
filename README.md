# eddie-brag

Staff-level professional accomplishment tracker for performance reviews.

## What This Is

A CLI tool for tracking work accomplishments with two workflows:
- **Daily capture** (`brag add`) — frictionless, just bucket/description/evidence
- **Weekly synthesis** (`brag enrich`) — add hours saved, business impact, strategic alignment

Data auto-syncs to a private GitHub repo. Never lose your career history.

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
# Last 7 days (default)
brag list

# Specific time periods
brag list --month 3
brag list --week 12
brag list --range 90d
brag list --all

# Generate review-ready report
brag report --month 3  # Grouped by bucket with totals
brag report --year 1   # Full first year summary

# Export to file
brag export --format csv --month 3
brag export --format json --all
```

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
  commands.go   - Command handlers
  prompts.go    - Interactive survey logic
  help.go       - Help text

internal/brag/
  brag.go       - Domain logic, storage, validation
  brag_test.go  - Unit tests

.github/workflows/
  test.yml      - CI (runs on macOS + Ubuntu)
```

Key design decisions:
- `time.Local` everywhere (no UTC drift on week/month boundaries)
- Self-healing `NextID` (protects against manual JSON edits)
- `EnrichedAt` timestamp (prevents infinite re-prompt loop)
- URL validation treats 401/403 as valid (internal tools are auth-protected)
- Auto git commit+push on every write (zero-effort backup)

## License

MIT
