---
id: gno-fcxvj
status: open
deps: []
links: []
created: 2026-04-27T01:17:12Z
type: feature
priority: 2
assignee: Stavros Korokithakis
---
# gn retag: rename a topic across all entries

Add 'gn retag <old> <new>' that replaces every occurrence of <old> with <new> in entry topic lists. Both names are normalized before comparison.

Scope: new internal/commands/retag.go using store.Update for atomic rewrite under exclusive lock.

Open decision: when an entry already has <new> as a topic, the rename collapses (dedupe) rather than producing duplicates. If <old> doesn't exist, error out. The new topic must satisfy the same length/format rules as on write.

Non-goals: regex/wildcard rename.

## Acceptance Criteria

gn retag old-topic new-topic updates all entries; entries that already had new-topic do not get duplicate entries. UpdatedAt is bumped only on entries that actually changed. Error if old topic does not exist or new topic is invalid.

