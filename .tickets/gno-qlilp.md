---
id: gno-qlilp
status: closed
deps: [gno-qrwtj]
links: []
created: 2026-04-24T08:39:49Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: show command

Implement 'gnosis show <target>' where <target> is either an entry ID (6 letters) or a topic name (anything else).

Entry ID: print the full entry — topics, text, related IDs (with their snippets if feasible), creation time. Format should be readable in a terminal, not JSON.

Topic: normalize the argument, find all entries whose normalized topics include it, print them in chronological order. Header line shows the topic's display form and entry count.

If target matches both (unlikely given ID format), prefer ID.

Output format: plain text with minimal decoration. Something like:

  abcxyz  [KeymasterTokenAuth, SessionManagement]  2026-04-24
  
  We decided on JWTs for backward compat with X. Conceptually opaque.
  
  Related: defuvw, ghijkm

Non-goals: paging, color, interactive navigation.

## Acceptance Criteria

gnosis show <id> prints one entry. gnosis show KeymasterTokenAuth prints all entries under that topic regardless of how the topic was originally cased.

