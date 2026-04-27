---
id: gno-mgfap
status: open
deps: []
links: []
created: 2026-04-27T01:16:22Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add gn doctor for knowledge base health

Add a 'gn doctor' command that audits the knowledge base and reports actionable issues. Checks should include: orphaned related links (pointing to deleted/missing entry IDs), entries with no topics, entries that have not been updated in a configurable timeframe (default 90 days), and potential duplicate entries (same or highly similar text). Output should be grouped by issue type with entry IDs and clear descriptions. Non-zero exit code if any issues are found, zero if clean.

## Acceptance Criteria

Running 'gn doctor' on a repo with known issues produces a structured report listing each issue type and affected entry IDs. A clean knowledge base exits with code 0 and no output.

