---
id: gno-toqzd
status: closed
deps: [gno-qrwtj]
links: []
created: 2026-04-24T08:40:02Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: topics command

Implement 'gnosis topics'. Lists all topics with entry counts, sorted by count descending (most-used first). Ties broken alphabetically on display form.

Output:

  12  KeymasterTokenAuth
   7  SessionManagement
   3  Billing
   …

The display form shown is the first-written one (already stored that way on entries).

Non-goals: filtering, search within topics list.

## Acceptance Criteria

After writing entries under a few topics, gnosis topics prints them with correct counts. Topics that normalize to the same key are collapsed into one line.

