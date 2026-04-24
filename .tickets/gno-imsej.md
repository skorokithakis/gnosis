---
id: gno-imsej
status: closed
deps: []
links: []
created: 2026-04-24T10:49:26Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# gnosis: search query handling + snippet format

Two search bugs from review:

1. Hyphenated topics break FTS5. Running 'gnosis search keymaster-token-auth' (the form shown by 'topics' and 'show') errors with 'no such column: token' because FTS5 parses hyphens as operators.
   Fix: in search command, sanitize the query before passing to FTS5. Simplest approach: if the query doesn't contain FTS5 operators the user likely meant, escape it as a phrase query by wrapping in double quotes, OR strip/replace hyphens with spaces. I lean: if the query looks like a bare term or set of terms (no AND/OR/NOT/parens/quotes), replace non-alphanumeric chars with spaces. Users who want advanced syntax can still use it explicitly.
   Alternative: quote each whitespace-separated term. Either way, the common path 'gnosis search keymaster-token-auth' must work.

2. FTS5's snippet() can include newlines from the indexed text, breaking the one-entry-per-line output contract. Scriptable output is important (agents pipe this).
   Fix: in search.go, collapse all whitespace runs (including newlines) to single spaces before printing the snippet.

## Acceptance Criteria

'gnosis search keymaster-token-auth' returns matching entries without error. Search output has exactly one line per hit even when entry text contains paragraph breaks.

