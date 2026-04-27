---
id: gno-wryiy
status: closed
deps: []
links: []
created: 2026-04-27T01:17:21Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn doctor: integrity audit

Add 'gn doctor' command that audits the JSONL store for issues and prints a report. Read-only by default; --fix flag could come later (out of scope here).

Checks (initial set, can grow):
- entries with Related IDs that don't exist (dangling refs)
- duplicate entry IDs
- entries with empty topics or empty text (corruption indicator)
- entries with topics that violate the 7-char minimum (legacy data from before validation)
- topics that exist on only one entry (informational, not an error)

Scope: new internal/commands/doctor.go. Exit code non-zero if any errors (not warnings) are found.

Open decision: distinction between 'error' and 'warning' levels.

## Acceptance Criteria

gn doctor prints a categorized report. Returns non-zero exit code when integrity errors exist. Healthy store prints 'no issues found' and exits 0.

