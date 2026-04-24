---
id: gno-dvvuv
status: closed
deps: []
links: []
created: 2026-04-24T10:49:29Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: show should fall back from ID to topic lookup

From review: common English words ('search', 'update', 'master', 'server', 'append') all match the 6-letter ID pattern, so 'gnosis show update' errors with 'entry not found' instead of showing entries tagged 'update'.

Fix: try ID lookup first; on miss, fall back to topic lookup. If both fail, error. This preserves the 'prefer ID' behavior for actual IDs while making short topic names usable.

## Acceptance Criteria

'gnosis show update' shows entries tagged 'update' when no entry with ID 'update' exists.

