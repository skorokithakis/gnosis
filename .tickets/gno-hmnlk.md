---
id: gno-hmnlk
status: open
deps: []
links: []
created: 2026-04-27T01:16:20Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Add shell completion scripts

Add a 'gn completion <shell>' command that generates shell completion scripts for bash, zsh, and fish. The command should output the appropriate script to stdout so users can redirect it to their shell configuration (e.g. 'gn completion bash > /etc/bash_completion.d/gn'). Completions should cover all existing subcommands and, where feasible, dynamic values like topic names and ID prefixes.

## Acceptance Criteria

A user can run 'gn completion bash|zsh|fish' and source the output to get working tab completions for gn subcommands.

