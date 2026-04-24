---
id: gno-rqvfh
status: closed
deps: [gno-gicyu]
links: []
created: 2026-04-24T16:58:26Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Colorize search output

Update internal/commands/search.go to use the termcolor helpers.

Before the result loop, build a []string of all entry IDs (already have entries loaded via store.ReadAll — reuse it; if the current code only builds a map, also collect the slice).

For each hit line:
- Replace the %-*s id padding with ColorizeID(hit.EntryID, allIDs), manually padded to idWidth spaces (use visible-length padding: pad after coloring so padding is not colored, compute visible width from the raw id length which is always 6).
- Colorize the topic with the topic color.
- Leave the snippet plain.

Keep the existing two-space column separators and overall layout.

## Acceptance Criteria

gn search output shows colored id (unique prefix bold magenta, rest dim) and cyan topics when stdout is a TTY; identical to current output when piped. Existing tests pass (buffer writer = no TTY = no color codes).

