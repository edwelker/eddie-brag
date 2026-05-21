# Future Integrations

Sprint 3+ work. External service connections and organizational features.

## Slack Integration

`brag search-mentions --id 4`

Searches Slack for:
- Mentions of Jira ticket ID
- Mentions of Confluence page URL
- Reactions/threads in relevant channels
- Auto-populates peer_recognition with links

## Jira Integration

`brag enrich-from-jira --id 8`

Fetches from Jira:
- Ticket creation date (estimate age of problem)
- Epic/label (suggests strategic_alignment)
- Linked issues (shows scope)
- Comments with "thanks" or reactions (peer_recognition)

## GitHub Import

`brag import-github --repos pma-api`

Auto-create entries from merged PRs.

## Calendar Integration

`brag import-calendar --date 2026-05-20`

Prompt based on meetings that day.

## Tag System

- Add tags during add: `brag add --tags "ci,test-infrastructure"`
- Report by theme: `brag report --theme "test-infrastructure"`
- Output shows progression: "3 entries over 6 weeks addressing test reliability"

## Skill Tagging

- `--skills python,testing,ci` flag on entries
- `brag report --skills testing` shows expertise across domain

## Workflow Automation

- Weekly reminders: `brag remind --weekly` (cron/launchd, Friday 5pm)
- Auto-generated digests: `brag digest --week 7` (email-formatted)
- End-of-day prompts

## External Sharing

- Portfolio export: `brag export --format portfolio` (HTML/PDF)
- Sanitized output: remove internal links, focus on outcomes
- Manager sharing: `brag share --manager` (quarterly update format)
