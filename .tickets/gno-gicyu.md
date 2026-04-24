---
id: gno-gicyu
status: closed
deps: []
links: []
created: 2026-04-24T16:58:16Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add color helper package

Add a new internal package (suggest internal/termcolor) that wraps github.com/fatih/color with semantic helpers for gnosis output.

Add dependency: github.com/fatih/color (run go get, commit go.mod/go.sum).

The package exposes:
- Functions to colorize semantic elements: ID unique prefix (bold magenta), ID remainder (dim/faint), topic (cyan), date (yellow), dim text (for labels like 'Related:'), bold (for counts).
- A function SplitUniquePrefix(id string, allIDs []string) (prefix, rest string) that returns the shortest prefix of id that does not prefix any other ID in allIDs, plus the remaining characters. If id is not in allIDs, treat it as present. If allIDs contains only id, the prefix is the first character.
- A convenience ColorizeID(id string, allIDs []string) string that returns the fully colorized ID string using the above.

fatih/color auto-disables color when stdout is not a TTY and honors NO_COLOR, so callers do not need TTY detection.

No behavior change to commands in this ticket; just the package + tests for SplitUniquePrefix covering: unique at 1 char, unique at full length (ambiguous shares 5-char prefix with another), single-entry case, id not in list.

## Acceptance Criteria

Package compiles, go test ./... passes, SplitUniquePrefix has unit tests for the four cases listed.

