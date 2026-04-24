---
id: gno-nsgxu
status: closed
deps: [gno-sjllx, gno-rphsl]
links: []
created: 2026-04-24T08:39:57Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: search command

Implement 'gnosis search <query>'. Ensures index is fresh, runs FTS5 query, prints ranked results.

Output: one line per hit — entry ID, primary (first) topic in display form, short snippet with matched terms highlighted if cheap (FTS5's snippet() function). Example:

  abcxyz  KeymasterTokenAuth    …decided on JWTs for backward compat…
  defuvw  SessionManagement     …session tokens expire after 24 hours…

Default limit: 20. --limit flag to override.

Query is passed to FTS5 mostly as-is; support FTS5's basic syntax (phrase queries, AND/OR). Don't try to be clever with preprocessing — let users lean on FTS5 when they want precision.

## Acceptance Criteria

Writing an entry then searching for a word from it returns the entry in results. Multi-word queries work. Topic names are searchable.

