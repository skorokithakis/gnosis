# gnosis

Your AI agents start from zero every session. gnosis gives them knowledge.

Add two lines to your AGENTS.md and every agent session searches existing memory before implementing, records decisions as it goes, and builds on what previous sessions learned. Decisions, rejected alternatives, constraints, and intent accumulate in a searchable knowledge base that lives in your repo and ships with your code.

## How it works

Agents interact through a small CLI (`gn`) that drives a three-phase loop on every task:

1. **Search** existing knowledge before implementing, so prior decisions, rejected alternatives, and known constraints surface before the agent writes any code.
2. **Record** decisions as they happen, while the reasoning is still fresh.
3. **Review** what should be captured after finishing, so nothing important falls through the cracks.

To enable this for every agent that touches your repo (Claude Code, Cursor, opencode, Aider) add two lines to your AGENTS.md:

```
At the start of any task, run `gn help plan` and follow its instructions.
After finishing a task, run `gn help review`.
```

Under the hood, entries are stored as JSONL in a `.gnosis` directory at your repo root, with a SQLite FTS5 index for full-text search.

## Why this matters

Code documents what a system does and documentation describes how to use it, but neither captures why things are the way they are, what was tried and abandoned, or what's known to be broken. This context lives in people's heads, gets lost in Slack threads, and evaporates when someone leaves the team or switches to a different project.

Humans have never been good at fixing this (they don't remember to write things down, and when they do it's inconsistent and eventually goes stale). Agents, on the other hand, follow instructions every time, which means they can reliably build up a knowledge base that grows with every session and stays useful without anyone having to maintain it.

## Installation

On macOS, install with Homebrew:

```sh
brew install --cask skorokithakis/tap/gnosis
```

On other platforms, download a pre-built binary from the [releases page](https://github.com/skorokithakis/gnosis/releases), or install with Go:

```sh
go install github.com/skorokithakis/gnosis/cmd/gn@latest
```

## See it in action

This repo uses gnosis on itself. Browse the [.gnosis directory](.gnosis/) to see the kind of knowledge that accumulates: decisions with their reasoning, rejected alternatives, known flaws and why they haven't been fixed yet, and intent about what's planned but not yet done.
