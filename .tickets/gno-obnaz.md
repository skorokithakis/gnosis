---
id: gno-obnaz
status: open
deps: []
links: []
created: 2026-04-27T01:17:15Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn search --ids-only

Add --ids-only flag to search that prints just one full entry ID per line, no topic, no snippet, no color. Enables piping: 'gn search foo --ids-only | xargs gn rm'.

Scope: small branch in internal/commands/search.go.

Note: print full IDs (not unique-prefix) so that downstream commands always work even if the entry set changes between pipe stages.

## Acceptance Criteria

gn search foo --ids-only emits one full 6-char ID per line, nothing else. Works with --limit. Empty result set produces no output.

