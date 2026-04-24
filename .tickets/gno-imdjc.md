---
id: gno-imdjc
status: closed
deps: []
links: []
created: 2026-04-24T10:49:37Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: doctrine and README text corrections

Stale text from before the normalization refactor and a README oversimplification:

1. internal/doctrine/help.txt, TOPICS section: 'Displayed in the form they were first written.' This is wrong post-refactor. Topics are displayed in normalized form. Update to: 'Displayed in normalized form (lowercase, dashes between words).'

2. README.md 'For AI agents' section currently says: 'Before starting work and after finishing code, run gnosis help and follow its instructions.' Reviewer #2 noted this doesn't match the three-touchpoint loop. Update to make the instruction accurate: the agent should run 'gnosis help' to read the protocol, then follow the three touchpoints it describes (search before implementing, write immediately for decisions, run 'gnosis help review' after finishing). Keep it short — pointing at gnosis help is fine, but don't misrepresent what's there.

3. help.txt line 37-38 says 'two touchpoints plus one rule for during work' then lists 3 items. Minor inconsistency. Either say 'three touchpoints' or restructure. Reviewer's call.

## Acceptance Criteria

help.txt topics section accurately reflects normalization. README 'For AI agents' section doesn't contradict the doctrine.

