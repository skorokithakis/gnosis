# gnosis

A knowledge base for what the code can't tell you.

Code documents what a system does, documentation describes how to use it, but there's a third kind of knowledge that neither captures: why things are the way they are, what was tried and abandoned, what's known to be broken, and what people intend to change but haven't yet. This context lives in people's heads, gets lost in Slack threads, and evaporates when someone leaves the team. gnosis gives it a home: a searchable, structured record of decisions, rejected alternatives, known pain points, and design intent that survives context switches and the passage of time.

## Why gnosis instead of ADRs, wikis, or code comments?

- **Searchable** — full-text search with stemming, not grep through markdown files.
- **Granular** — one entry per decision, not one document per quarter.
- **Agent-native** — AI coding agents can search existing knowledge before implementing and record decisions as they go, so context survives across sessions.
- **No infrastructure** — entries are JSONL files in your repo. No server, no database to manage, no sync to configure. It ships with your code in git.

## Installation

Download a pre-built binary from the [releases page](https://github.com/skorokithakis/gnosis/releases), or install with Go:

```sh
go install github.com/skorokithakis/gnosis/cmd/gn@latest
```

## Quick start

```sh
# Record a decision
gn write auth,tokens "We use short-lived JWTs rather than sessions because
the service is stateless by design. Sessions were considered and rejected
due to the complexity of distributed session storage."

# Before touching an area, search for existing knowledge
gn search auth token expiry

# Show everything recorded under a topic
gn show auth

# See all topics and how many entries each has
gn topics
```

## For AI coding agents

gnosis is designed to work with AI coding agents. Agents lose context between sessions — they don't know why you chose Postgres over DynamoDB, why that endpoint is deliberately slow, or what was tried and failed last month. gnosis gives them a way to check before they redo your work.

Add this to your agent's system prompt or AGENTS.md:

```
At the start of any task, run `gn help plan` and follow its instructions.
After finishing a task, run `gn help review`.
```

This gives the agent a three-step loop: search existing knowledge before implementing, write entries when decisions are made, and review what should be recorded after finishing.
