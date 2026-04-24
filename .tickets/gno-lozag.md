---
id: gno-lozag
status: closed
deps: [gno-drkdd]
links: []
created: 2026-04-24T09:21:28Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: write README

Write a README.md for the repository. Should cover: what gnosis is (one paragraph — use the PURPOSE section from help.txt as a starting point but make it a proper README intro, not a copy-paste), installation (go install github.com/skorokithakis/gnosis/cmd/gnosis@latest), quick start (gnosis help, gnosis write, gnosis search), and a pointer to 'gnosis help' for full documentation. Keep it short — the tool's own help text is the real docs. Non-goals: badges, contributing guide, changelog (release-please handles that).


## Notes

**2026-04-24T10:34:51Z**

Update: README should include a clear 'For AI agents' section describing what to add to the agent's system prompt. Something like: 'Add the following to your agent config: Before starting work and after finishing code, run gnosis help and follow its instructions.' Keep this section prominent — it's the main onboarding point for the tool's primary use case.
