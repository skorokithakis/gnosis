---
id: gno-ahqqj
status: closed
deps: [gno-qrwtj]
links: []
created: 2026-04-24T08:40:20Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: rm command

Implement 'gnosis rm <id> [<id>...]'. Removes entries.

Behavior:
- Load all entries, filter out the requested IDs, rewrite JSONL atomically under the file lock.
- Error (and do nothing) if any ID doesn't exist — partial removal is confusing.
- Warn if any surviving entry has the removed ID in its Related list, but don't modify those entries automatically. Just print something like 'note: entry abcxyz references removed entry defuvw'.
- Print removed IDs on success.

Non-goals: undo, confirmation prompt (the agent invokes this; prompting breaks automation; human users can re-add from git).

## Acceptance Criteria

Removing an existing entry updates the JSONL. Removing a non-existent ID errors without modifying the file. Dangling Related references are reported but not cleaned up.

