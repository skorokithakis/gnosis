---
id: gno-vaclb
status: closed
deps: [gno-icdbl]
links: []
created: 2026-04-24T08:40:47Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: help command + doctrine text

Implement 'gnosis help' (and bare 'gnosis'). Prints the doctrine — the spirit of the tool, what to record, what not to record, the agent's before/after loop, and a terse subcommand listing.

The doctrine text lives as a markdown file embedded via Go's embed package (e.g. internal/doctrine/help.md). Editing the doctrine is editing a file, not a string literal.

The text must cover, at minimum:
- Purpose: why this tool exists. Agents and humans forget the planning reasoning; code records the what, not the why. This captures the intangibles that only live in people's heads.
- What to record: decisions taken AND decisions explicitly rejected, known-buggy areas, 'we've been meaning to redesign X', constraints that shaped a design, intent behind a structure. Any time someone says 'we considered X but chose Y' or 'this is flaky' or 'we want to rework this' — that's a candidate.
- What NOT to record: anything visible in the code or existing docs. If a code reader would learn it, don't duplicate it. The value is negative space and intent.
- The agent loop:
  - BEFORE starting work: 'gnosis search <keywords>' on the task's topic. Read what exists. Flag conflicts with recorded decisions to the user before proceeding — they may still be right, may be stale, or may need explicit supersession.
  - AFTER finishing: identify intangibles from the session (decisions made, alternatives rejected, pain points mentioned, planned reworks) and 'gnosis write' them. Capture only reasoning that would be lost otherwise.
- Tidying: stale KB is worse than small KB. When a recorded decision is overturned, rm the obsolete entry or write a new one that references the old ID in prose.
- Topics: free-form tags, case/separator-insensitive on lookup, CamelCase/kebab-case/snake_case all collapse. Topics spring into existence on first write. Multi-topic entries are fine (comma-separated).
- Brief command reference at the end.

Tone: direct, opinionated, no filler. The text IS the product — it's what shapes agent behavior. Must be reviewed carefully before closing.

I (the architect) will draft this text and show it to the user for review before marking this ticket done. Expect at least one revision round.

Non-goals: per-command --help output (that's stdlib flag's job).

## Acceptance Criteria

'gnosis' and 'gnosis help' print the full doctrine. Text covers purpose, what-to-record, what-not-to-record, the before/after loop, tidying, topics model, and a command reference. User has reviewed and approved the final text.


## Notes

**2026-04-24T09:21:17Z**

Design update: Doctrine text files are already written at internal/doctrine/help.txt and internal/doctrine/review.txt. Embed them via Go's embed package. 'gnosis help' prints help.txt, 'gnosis help review' prints review.txt. No args and 'help' are equivalent.
