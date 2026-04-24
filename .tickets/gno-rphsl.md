---
id: gno-rphsl
status: closed
deps: [gno-qrwtj]
links: []
created: 2026-04-24T08:39:41Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: write command

Implement 'gnosis write <topics> <text> [--related id1,id2]'.

<topics>: comma-separated list, stored in display form as given (no normalization of the display string beyond trimming whitespace). Duplicates (same normalized form) are collapsed, first-written display wins.

<text>: the entry body. If not provided on the command line and stdin is not a TTY, read from stdin. (Supports 'echo ... | gnosis write foo'.)

--related: optional comma-separated entry IDs. Validated — error if any referenced ID doesn't exist.

Behavior: generate ID, append to JSONL, print the new ID (and display-form topic:ID ref for human use) to stdout. Do not invoke the indexer synchronously — next search will refresh it. Keep it fast.

Validation:
- At least one topic required.
- Text must be non-empty after trimming.
- Topics must not contain commas (already split) or be empty strings.

Non-goals: interactive prompts, editor invocation, --related free-text search.

## Acceptance Criteria

gnosis write foo,Bar 'hello' prints a 6-char ID. Entry appears in JSONL with both topics in display form. gnosis write Foo 'second' reuses the existing 'foo' topic's normalized key, preserves original 'foo' display (first-write wins).

