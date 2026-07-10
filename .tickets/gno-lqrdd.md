---
id: gno-lqrdd
status: closed
deps: []
links: []
created: 2026-07-10T21:44:28Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Harden gn write argument handling

Reject more than one positional text argument with a clear error suggesting shell quoting. Preserve stdin fallback and flexible --related placement. Add focused command tests. Non-goal: joining positional arguments or changing documented command syntax.

## Acceptance Criteria

Extra positional arguments are rejected rather than silently discarded; existing write behavior remains unchanged.


## Notes

**2026-07-10T21:45:14Z**

ready for implementation
