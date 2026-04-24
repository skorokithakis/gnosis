---
id: gno-kishl
status: closed
deps: []
links: []
created: 2026-04-24T23:36:13Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add word-wrap helper

Add a small word-wrap helper used by gn show. New package internal/textwrap (or a private helper inside internal/commands — pick whichever is cleaner) exposing something like Wrap(s string, width int) string.

Behavior:
- width <= 0 returns s unchanged (used for non-TTY output).
- Splits input on existing newlines and wraps each line independently, so paragraphs/blank lines are preserved.
- Within a line, word-wraps on whitespace: greedily fills up to width columns, breaks before a word that would exceed it.
- A single token longer than width overflows on its own line rather than being broken mid-token.
- Operates on rune count, not byte count.
- Does not need to account for ANSI escape sequences (body text has none).

Include unit tests covering: no-op when width<=0, basic wrap, preserved blank lines, oversized-token overflow, multi-paragraph input, unicode width counted by runes.

## Acceptance Criteria

Pure function with tests; no other files touched.

