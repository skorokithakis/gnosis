# Product strategy

Last updated: 2026-04-25

## What gnosis is

gnosis is agent memory infrastructure. It gives AI coding agents a way to accumulate knowledge across sessions so that each session builds on the last instead of starting from zero. Agents are the primary writers and readers; humans read entries when they need to but are not expected to maintain the knowledge base.

It is not a knowledge management tool, not an ADR replacement, and not a wiki. Those framings put gnosis in competition with tools that have been failing to solve the "write things down" problem for decades. gnosis sidesteps that problem entirely by making agents do the writing.

## Target user

Teams that use AI coding agents for development. The adopter is a developer who adds gnosis instructions to their repo's AGENTS.md. From that point on, every agent session that touches the repo participates automatically.

## Distribution

gnosis spreads through repos, not through marketing. The adoption path:

1. One developer installs gnosis and adds two lines to AGENTS.md.
2. Agents start writing entries. The `.gnosis` directory appears in the repo.
3. Other contributors' agents pick up the AGENTS.md instructions and start searching/writing too.
4. The knowledge base becomes useful enough that removing it would be a loss.

This means gnosis doesn't require org-wide buy-in or a procurement process. It enters through a single commit.

## Competitive landscape

The real competition is "nothing." Most teams don't record decisions, rejected alternatives, or institutional knowledge in any structured way. The tools that exist (ADRs, wikis, Notion docs) require human discipline to maintain, which is why they go stale. gnosis competes by removing the human from the write path.

## Key risk

Signal-to-noise in agent-written entries. If agents write too much low-value content, the knowledge base becomes noisy and people delete `.gnosis`. If agents write too little, there's nothing to search and the tool appears useless. The doctrine (`gn help plan`, `gn help review`) is the current mechanism for guiding agent behavior, and improving it is an ongoing priority.

## Open questions

- **Signal-to-noise optimization.** What makes an agent write a good entry vs. a useless one? Is the current doctrine sufficient, or does it need structural changes (entry templates, validation, quality scoring)?
- **Cross-repo knowledge.** Should gnosis support knowledge that spans multiple repos (org-wide decisions, shared conventions)? Not yet, but worth tracking as a future direction.
- **Entry lifecycle.** Do entries go stale? Should there be a mechanism for agents to flag or retire outdated entries?
