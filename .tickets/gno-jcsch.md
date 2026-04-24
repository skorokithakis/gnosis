---
id: gno-jcsch
status: closed
deps: []
links: []
created: 2026-04-24T16:44:26Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Enforce 7-char topic minimum

In write and edit, reject any topic whose normalized form is shorter than 7 characters. Error message should name the offending topic.

This prevents collision between topic names and ID prefixes (which are <=6 chars). Existing entries with short topics are not migrated or rewritten.

## Acceptance Criteria

gn write and gn edit refuse entries containing a topic that normalizes to fewer than 7 chars, with a clear error. Existing short topics in storage remain as-is.

