---
id: gno-xhtoz
status: closed
deps: []
links: []
created: 2026-04-24T16:53:36Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Review fixes: empty prefix, rune count, edit buffer Related

Three follow-ups from the final review of the ID-prefix feature:

1. ResolveIDPrefix should reject an empty prefix string (return not found). Currently an empty prefix matches every entry's ID and, in a one-entry store, resolves to that entry. Guard at the top of internal/commands/resolve.go.

2. Use rune count (utf8.RuneCountInString or len([]rune(...))) instead of len() for the length comparisons that represent 'characters':
   - internal/commands/show.go dispatch threshold (<=6 / >=7)
   - internal/commands/write.go short-topic check
   - internal/commands/edit.go short-topic check (normalizeAndDeduplicateTopics)

3. In gn edit, the Related list parsed from the edit buffer must be resolved via ResolveIDPrefix (same as write --related), so the stored Related list always contains full IDs. Currently validateEditedEntry only checks exact-match existence. Resolve each parsed related value against the locked snapshot (excluding the entry being edited), store the resolved full IDs on the updated entry, and surface ambiguous/not-found errors.

Add targeted tests for each of the three fixes.

## Acceptance Criteria

ResolveIDPrefix(..., "") returns a not-found error. Multibyte-character topics and queries are measured in runes for the 7-char rule. gn edit accepts ID prefixes in the buffer's Related line and stores full IDs. go test ./... passes.

