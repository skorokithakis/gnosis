---
id: gno-yphrz
status: closed
deps: []
links: []
created: 2026-07-10T21:44:28Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Remove expired stale-lock compatibility

Remove the post-2026-07-01 stale repository-lock cleanup path, its test-only exports, and its obsolete tests. Retain current cache-directory locking and its tests. Non-goal: redesigning locking or concurrency behavior.

## Acceptance Criteria

No stale-lock compatibility symbols or obsolete tests remain; current lock behavior is preserved.


## Notes

**2026-07-10T21:45:14Z**

ready for implementation
