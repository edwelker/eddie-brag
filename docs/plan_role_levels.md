# Role-Level Contextual Prompts

Staff engineer expectations differ from IC or Manager. Tool should be role-aware.

## Schema Addition

```go
type BragDocument struct {
    RoleTitle     string    `json:"role_title"`
    RoleLevel     string    `json:"role_level"` // "IC", "Senior", "Staff", "Principal", "Manager"
    RoleStartDate time.Time `json:"role_start_date"`
    NextID        int       `json:"next_id"`
    Entries       []Entry   `json:"entries"`
}
```

Set during `brag init` or `brag config role-level Staff`.

## Role-Specific Enrichment Requirements

| Role | Required Fields | Validation |
|------|----------------|------------|
| IC/Mid | Evidence, Description | Basic completeness |
| Senior | + Hours saved (for Process) | Quantified impact |
| Staff | + Business metric, Strategic alignment | Org-level thinking |
| Principal/Staff+ | + Peer recognition (all buckets) | Multiplier effect |
| Manager | + Team size affected | People leadership |

## Role-Specific Prompt Examples

IC/Mid-level (Delivery bucket):
```
Examples:
  - Completed feature X
  - Fixed bug affecting Y users
  - Deployed Z service
```

Staff-level (Delivery bucket):
```
Examples:
  - Shipped feature enabling $X ARR pipeline
  - Resolved P0 blocker affecting X customers, prevented churn
  - Architectural decision unblocked 3 teams for 2 quarters
```

## Role-Specific Bucket Weighting

Staff+ warnings:
- Too many Delivery entries, not enough Architecture/Leadership
- "Staff engineers are expected to show impact beyond individual execution."

## Role-Specific Report Templates

`brag summary --month 1` differs by role:

IC: "Completed 5 features, fixed 12 bugs, participated in on-call rotation."

Staff: "Led test infrastructure improvements saving 40 business days annually. Unblocked QA team after 4-month outage. Proposed security automation workflow adopted by platform team."
