---
id: gno-onwhm
status: closed
deps: [gno-vaclb, gno-rphsl, gno-nsgxu]
links: []
created: 2026-04-24T08:40:56Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: update agent prompt to point at gnosis

Replace the .knowledge/ section in the user's agent configuration with a short pointer to 'gnosis help'.

Files likely affected:
- ~/.config/opencode/AGENTS.md (personal global conventions) — only if it contains KB instructions; check first.
- The architect system prompt (wherever the user maintains it) — has a large '.knowledge/' section that must be replaced with a one-liner.

Replacement text (draft, user reviews before applying):

  Knowledge base
  Before starting work and after finishing, run 'gnosis help' and follow its instructions.

Ask the user where the architect prompt lives before editing it. Do not edit blindly.

Non-goals: migrating any existing .knowledge/ content — none exists yet. If it exists later, that's a separate task.

## Acceptance Criteria

Agent prompt no longer contains the verbose .knowledge/ section. The replacement pointer is in place. User has confirmed the new text reads correctly.


## Notes

**2026-04-24T10:34:58Z**

Cancelled — user will update their agent prompt themselves using the README as a guide.
