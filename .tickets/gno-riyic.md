---
id: gno-riyic
status: closed
deps: []
links: []
created: 2026-04-27T01:17:01Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn show --expand-related

When showing an entry, optionally inline the text of all entries in its Related list, one level deep only. Cycles are not possible since we only go one level.

Scope: extend internal/commands/show.go and main.go.

Default proposal: opt-in via --expand-related flag (don't change current default to avoid surprise). Format: after the main entry body, print '--- related ---' separator and each related entry with a slightly indented or dimmed header.

Non-goals: multi-level traversal, graph visualization.

## Acceptance Criteria

gn show <id> --expand-related prints the entry plus full text of each related entry. Works with topic-mode show too (related entries from each matched entry). Without the flag, output is unchanged.

