---
id: gno-ukhpc
status: closed
deps: []
links: []
created: 2026-04-24T16:44:20Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# ID prefix resolver and wiring

Add a prefix-based ID resolver and use it at every site that accepts an entry ID: show, rm, edit, and write --related.

Resolver: given all entries and a prefix string, return the single full ID whose prefix matches, or an error. Errors:
- not found (no entry ID starts with the prefix, or prefix contains chars outside the ID alphabet)
- ambiguous (list the candidate full IDs in the error message)

Minimum prefix length is 1. A full 6-char ID is just the degenerate case of a unique prefix.

Wiring:
- show: dispatch purely on query length. <=6 chars -> ID-prefix lookup only. >=7 chars -> topic lookup only. Remove the existing ID->topic fallback. Replace/remove isEntryID.
- rm: resolve each argument via the resolver. Keep the current 'validate all before mutating' behavior; if any arg fails to resolve, refuse the whole operation.
- edit: resolve the single arg via the resolver before loading.
- write --related: resolve each comma-separated value via the resolver; the stored Related list must contain the full IDs.

Resolution is silent; downstream output continues to print the full resolved ID as today.

## Acceptance Criteria

gn show/rm/edit/write --related all accept unique ID prefixes of any length >=1. Ambiguous prefixes error with candidate list. show no longer falls back between ID and topic lookup; the split is by query length.

