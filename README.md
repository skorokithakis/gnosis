# gnosis

A knowledge base for what the code can't tell you.

Code documents what a system does, documentation describes how to use it, but there's a third kind of knowledge that neither captures: why things are the way they are, what was tried and abandoned, what's known to be broken, and what people intend to change but haven't yet. This context lives in people's heads, gets lost in Slack threads, and evaporates when someone leaves the team. gnosis gives it a home: a searchable, structured record of decisions, rejected alternatives, known pain points, and design intent that survives context switches and the passage of time.

## Installation

```
go install github.com/skorokithakis/gnosis/cmd/gn@latest
```

Requires Go 1.25 or later. Release binaries are available on the [Releases page](https://github.com/skorokithakis/gnosis/releases).

## Quick start

```sh
# Record a decision
gn write auth,tokens "We use short-lived JWTs rather than sessions because the service is stateless by design. Stateful sessions were considered and rejected."

# Search for relevant knowledge before touching an area
gn search auth token expiry

# Show everything recorded under a topic
gn show auth

# See all topics and how many entries each has
gn topics
```

## For AI agents

> **Add this to your agent's system prompt:**
>
> ```
> At the start of any task, run `gn help plan` and follow its instructions.
> After finishing a task, run `gn help review`.
> ```

`gn help plan` describes three touchpoints: searching existing knowledge before implementing, writing entries immediately when decisions are made, and reading `gn help review` after finishing. The second command in the snippet above is a belt-and-suspenders reminder so the review criteria are fresh in context right when the agent needs them.

## Full documentation

Run `gn help plan` for the full doctrine, and `gn help review` for guidance on recording knowledge after a session. `gn help` by itself prints a short overview and the command list.

## License

MIT — Copyright (c) 2026 Stavros Korokithakis
