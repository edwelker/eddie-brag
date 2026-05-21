# Phase 2, Sprint 2: Impact Amplification

Medium effort, high value. Focus: quantify and narrate impact.

## 5. Narrative Summary Generation (`brag summary`)

`brag summary --month 1` (or --week, --range, --all)

Template-based prose output:
```
## Impact Summary: Month 1

Primary focus: {most common bucket}
Key metrics: {total hours saved}, {entry count} accomplishments, {team members affected}

Strategic initiatives:
- {bucket}: {entry count} entries ({hours saved})
  - {top 3 entries by impact}

Evidence of impact:
- {peer recognition links}
- {business metrics achieved}
```

## 6. Team Size Impact Multiplier

Schema addition:
```go
TeamSize        *int    `json:"team_size,omitempty"`
FinancialImpact *string `json:"financial_impact,omitempty"`
```

During enrich for Process/Leadership buckets:
- "How many people does this affect?"
- Options: "Just me (1)", "My team (~10)", "Multiple teams (~50)", "Whole org (100+)"

Calculations:
- `hours_saved * team_size = person_hours`
- Display: "321 hrs x 10 people = 3,210 person-hours = 401 person-days"
- Financial: `person_hours * $100/hr`

## 7. Missing Evidence Validation

Warning + confirmation on add without evidence:
```
Warning: Evidence missing. This entry will be harder to verify during reviews.

Options:
  1. Add evidence URL now
  2. Add placeholder (e.g., "Slack thread - need link")
  3. Skip (will prompt during brag review)
```

## Verification

- Process entry prompts for team size, displays "321 hrs x 10 people = $321K value"
- `brag summary --month 1` generates prose summary with strategic narrative
- Adding entry without evidence triggers warning + confirmation
