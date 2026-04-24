# gnosis

A knowledge base for what the code can't tell you.

Code documents what a system does. Documentation describes how to use it. But there's a third kind of knowledge that neither captures: why things are the way they are, what was tried and abandoned, what's known to be broken, and what people intend to change but haven't yet. This context lives in people's heads, gets lost in Slack threads, and evaporates when someone leaves the team. gnosis gives it a home — a searchable, structured record of decisions, rejected alternatives, known pain points, and design intent that survives context switches and the passage of time.

## Installation

```
go install github.com/skorokithakis/gnosis/cmd/gnosis@latest
```

Requires Go 1.25 or later. Release binaries will be available on the [Releases page](https://github.com/skorokithakis/gnosis/releases) once the first release is cut.

## Quick start

```sh
# Record a decision
gnosis write auth,tokens "We use short-lived JWTs rather than sessions because the service is stateless by design. Stateful sessions were considered and rejected."

# Search for relevant knowledge before touching an area
gnosis search auth token expiry

# Show everything recorded under a topic
gnosis show auth

# See all topics and how many entries each has
gnosis topics
```

## For AI agents

> **Add this to your agent's system prompt:**
>
> ```
> At the start of any session, run `gnosis help` and follow its instructions. After
> finishing a task, run `gnosis help review`.
> ```

`gnosis help` is the entry point — it describes three touchpoints: searching existing knowledge before implementing, writing entries immediately when decisions are made, and reading `gnosis help review` after finishing. The second command in the snippet above is a belt-and-suspenders reminder so the review criteria are fresh in context right when the agent needs them.

## Full documentation

Run `gnosis help` for the full doctrine, and `gnosis help review` for guidance on recording knowledge after a session.

## License

MIT — Copyright (c) 2026 Stavros Korokithakis
