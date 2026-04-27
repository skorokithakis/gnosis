---
id: gno-aruyl
status: closed
deps: []
links: []
created: 2026-04-27T01:16:51Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn latest: show recent entries

Add 'gn latest [--limit N]' command that prints entries in reverse-chronological order by entry timestamp (newest first). Output format matches 'gn search' (id, topics, snippet).

Scope: new file internal/commands/latest.go, dispatch in cmd/gn/main.go, tests next to it.

Non-goals: filtering by topic/date, pagination beyond --limit.

## Acceptance Criteria

gn latest prints newest entries first; --limit N restricts count; default limit is 20. Empty store prints nothing and exits 0.
