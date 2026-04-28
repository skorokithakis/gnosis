---
id: gno-wpfjr
status: closed
deps: []
links: []
created: 2026-04-28T22:34:46Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add updated-date column to latest and search output

Both 'gn latest' and 'gn search' currently print '<id>  <topic>  <snippet>' per row, with no indication of when the entry was last touched. Add a date column between the id and topic columns showing entry.UpdatedAt formatted as '2006-01-02', colored via termcolor.Date (consistent with how 'show' already renders dates).

Files: internal/commands/latest.go, internal/commands/search.go, plus their _test.go files.

The date column has a fixed 10-char width, so no max-width bookkeeping is needed for it. Keep the existing two-space separator between every column. Update the affected tests to match the new output.

Non-goals: no changes to 'show' (it already displays both created and updated). No flag to hide the date. No relative-time formatting.

