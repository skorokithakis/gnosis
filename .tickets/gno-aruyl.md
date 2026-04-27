---
id: gno-aruyl
status: open
deps: []
links: []
created: 2026-04-27T01:16:51Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn list: show recent entries

Add 'gn list [--limit N]' command that prints entries in reverse-chronological order (newest first). Output format should match 'gn search' (id, primary topic, snippet) so results can be piped into other gn commands.

Scope: new file internal/commands/list.go, dispatch in cmd/gn/main.go, tests next to it.

Non-goals: filtering by topic/date (that's separate work), pagination beyond --limit.

## Acceptance Criteria

gn list prints newest entries first; --limit N restricts count; default limit is 20 (matches search). Empty store prints nothing and exits 0.

