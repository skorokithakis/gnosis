---
id: gno-newbq
status: closed
deps: [gno-sjllx]
links: []
created: 2026-04-24T08:40:25Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: reindex command

Implement 'gnosis reindex'. Forces a full rebuild of the SQLite FTS5 index from the current JSONL. Mostly a convenience for when the cache gets confused or after manual JSONL editing (git merges, etc).

Output: a line reporting N entries indexed and where the index lives.

Non-goals: partial/incremental rebuild, index statistics.

## Acceptance Criteria

gnosis reindex rebuilds the index; a subsequent search returns expected results.

