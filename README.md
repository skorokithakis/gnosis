# gnosis

Have you ever spent a day on an approach, only to find out that a colleague already tried it and rejected it, for a reason nobody wrote down?

AI agents do this on every single task, because every session starts from zero. The agent can read the code and the docs, and it's very good at that, but it has no idea that I tried the obvious approach last quarter and it fell over, or that a feature is half-built on purpose, or that a dependency is pinned because the vendor broke something in the next version. That knowledge lives in my head and in a few chat threads, so the agent re-derives it (slowly), or re-tries the thing that didn't work (worse), or quietly reverses a decision I made for a reason it couldn't see.

gnosis is a small CLI, `gn`, that gives agents a place to keep that knowledge, and a set of instructions that make them actually use it. It records the "why" of a codebase: decisions, the alternatives that were considered and rejected, constraints from outside the code, and things that are known to be broken and why nobody has fixed them yet. Entries live in a `.gnosis` directory in your repo, as JSONL with a SQLite full-text index next to it, so they travel with the code and show up in the diff like everything else.

## What it looks like

This repo uses gnosis on itself. Here is a real entry from it:

> Considered replacing SQLite FTS5 with Bleve for full-text search. Rejected. The main argument for Bleve would be dropping cgo, but we already use modernc.org/sqlite (pure-Go), so static binaries and cross-compilation are not at risk. FTS5 gives us BM25 ranking, porter stemming, unicode61 tokenization, prefix indexes (2/3/4), and snippet() with custom delimiters in ~260 lines and a single-file index that's trivially disposable/rebuildable from the JSONL source of truth. Bleve would add ~30+ transitive deps, a larger binary, a directory-shaped index, and require re-implementing the snippet sentinel trick. [...] Revisit only if real demand for typo-tolerant search emerges.

That is the part I care about most. Code records what you built, and a decent commit message records why, but nothing records what you *didn't* build, because there's no code to attach the reason to. The next agent (or the next person) that thinks "why not Bleve?" will search, find this, and either move on or come back with an argument the entry doesn't cover. Either way, it starts from where I left off instead of from nothing.

The same goes for the other kinds of entries: the release job has a known failure mode that nobody has fixed because nobody has hit it yet, and there is no MCP server because the agents that matter all have shell access. Both are written down, with the reasoning, in [`.gnosis/`](.gnosis/).

## Setup

Install the binary (see below), then add two lines to your `AGENTS.md`, or `CLAUDE.md`, or whatever file your agent reads at the start of a session:

```
At the start of any task, run `gn help plan` and follow its instructions.
After finishing a task, run `gn help review`.
```

That's it. The doctrine the agent reads tells it to search before it plans, to write entries as decisions happen rather than at the end, and to review the session afterwards for anything it should have recorded. It also spends a lot of words on what *not* to record, which matters more than it sounds: anything you could learn by reading the code doesn't belong here, because duplication rots and stale knowledge is worse than none.

Agents are good at this and humans aren't, and I'm including myself in "humans". Nobody keeps a decision log going, but an agent follows the instruction every single time, and it has the context fresh in its window at the moment the decision is made, which is the only moment anyone ever has the energy to write it down.

## Installation

On macOS, install with Homebrew:

```sh
brew install --cask skorokithakis/tap/gnosis
```

On other platforms, download a pre-built binary from the [releases page](https://github.com/skorokithakis/gnosis/releases), or install with Go:

```sh
go install github.com/skorokithakis/gnosis/cmd/gn@latest
```

## Commands

The agent learns these from `gn help`, so you mostly won't type them yourself, but they are useful for browsing what has accumulated:

```
gn write <topics> <text>   Record an entry. Topics are comma-separated and created on first use.
gn search <query>          Full-text search. "auth OR token" for either, "auth NOT token" to exclude.
gn show <id|topic>         Show one entry, or every entry under a topic.
gn latest                  The most recent entries.
gn topics                  All topics with entry counts.
gn edit <id>               Edit an entry in $EDITOR.
gn rm <id>                 Remove an entry.
```

Entries have six-letter IDs, and any unique prefix works where an ID is expected.

## Downsides

Search is keyword search (SQLite FTS5 with stemming), not semantic search, so the agent has to guess roughly the right words. The doctrine tells it to OR its terms together, which helps, but it will miss things sometimes.

Entries are only as good as the agent's judgement about what is worth recording, and that judgement is not perfect. The review step catches some of the noise, but you should expect to prune the odd entry that just restates the code.

It's a CLI, so it only works with agents that have shell access. Claude Code, Cursor, opencode and Aider all do. There is no MCP server, for the reasons recorded in `.gnosis`.

That's it for downsides, really.

If you want to see what a few months of this produces, the [`.gnosis` directory](.gnosis/) in this repo is the whole knowledge base for gnosis itself, and I think it tells a new agent more about the project than the git log does.
