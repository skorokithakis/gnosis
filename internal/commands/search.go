package commands

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// fts5OperatorPattern matches any character sequence that indicates the user
// is intentionally using FTS5 query syntax rather than typing a bare search
// term. We look for: double-quote (phrase query), parentheses (grouping),
// colon (column qualifier), and the uppercase boolean keywords AND/OR/NOT
// surrounded by spaces (FTS5 only treats these as operators when uppercase).
var fts5OperatorPattern = regexp.MustCompile(`"|[():]| AND | OR | NOT `)

// nonAlphanumericPattern matches one or more consecutive characters that are
// neither alphanumeric nor whitespace. These are replaced with a single space
// when sanitizing a bare query so that e.g. "keymaster-token-auth" becomes
// "keymaster token auth", which FTS5 interprets as an AND-of-three-terms match.
var nonAlphanumericPattern = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

// whitespaceRunPattern matches one or more consecutive whitespace characters,
// including newlines. Used to collapse multi-line FTS5 snippets to a single line.
var whitespaceRunPattern = regexp.MustCompile(`\s+`)

// sanitizeQuery returns the query unchanged if it contains FTS5 operator
// syntax (the user knows what they are doing), or replaces runs of
// non-alphanumeric, non-whitespace characters with spaces so that bare
// hyphenated terms like "keymaster-token-auth" do not trigger FTS5's column
// qualifier parser.
func sanitizeQuery(query string) string {
	if fts5OperatorPattern.MatchString(query) {
		return query
	}
	return nonAlphanumericPattern.ReplaceAllString(query, " ")
}

const defaultSearchLimit = 20

// Search implements the "gnosis search <query> [--limit N]" command. It opens
// the FTS5 index, ensures it reflects the current JSONL state, runs the query,
// and prints one line per result. argv should be os.Args[2:] (everything after
// "search"). Output is written to writer so callers and tests can capture it.
func Search(store *storage.Store, argv []string, writer io.Writer) error {
	limit := defaultSearchLimit
	var queryArgs []string

	for i := 0; i < len(argv); i++ {
		if argv[i] == "--limit" {
			if i+1 >= len(argv) {
				return fmt.Errorf("--limit requires a value")
			}
			parsed, err := strconv.Atoi(argv[i+1])
			if err != nil || parsed <= 0 {
				return fmt.Errorf("--limit value must be a positive integer, got %q", argv[i+1])
			}
			limit = parsed
			i++
		} else {
			queryArgs = append(queryArgs, argv[i])
		}
	}

	if len(queryArgs) == 0 {
		return fmt.Errorf("usage: gnosis search <query> [--limit N]")
	}

	// Join all non-flag arguments as the query so that users can write
	// "gnosis search hello world" without quoting.
	query := strings.Join(queryArgs, " ")

	query = sanitizeQuery(query)

	repoRoot, err := storage.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	idx, err := index.Open(repoRoot, store)
	if err != nil {
		return fmt.Errorf("opening index: %w", err)
	}
	defer idx.Close()

	if err := idx.EnsureFresh(); err != nil {
		return fmt.Errorf("refreshing index: %w", err)
	}

	hits, err := idx.Search(query, limit)
	if err != nil {
		return fmt.Errorf("searching: %w", err)
	}

	if len(hits) == 0 {
		return nil
	}

	// Build a map from entry ID to entry so we can look up the primary topic
	// for each hit without a linear scan per result.
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}
	entryByID := make(map[string]storage.Entry, len(entries))
	for _, entry := range entries {
		entryByID[entry.ID] = entry
	}

	// Determine column widths for readable alignment. We pad the ID and topic
	// columns so that snippets start at a consistent horizontal position.
	const idWidth = 6
	maxTopicWidth := 0
	for _, hit := range hits {
		entry, found := entryByID[hit.EntryID]
		if !found {
			continue
		}
		primaryTopic := primaryTopicOf(entry)
		if len(primaryTopic) > maxTopicWidth {
			maxTopicWidth = len(primaryTopic)
		}
	}

	for _, hit := range hits {
		entry, found := entryByID[hit.EntryID]
		primaryTopic := ""
		if found {
			primaryTopic = primaryTopicOf(entry)
		}
		// Collapse all whitespace runs (including newlines that FTS5's
		// snippet() may include from the indexed text) to single spaces so
		// that each result occupies exactly one output line.
		snippet := whitespaceRunPattern.ReplaceAllString(hit.Snippet, " ")
		fmt.Fprintf(writer, "%-*s  %-*s  %s\n",
			idWidth, hit.EntryID,
			maxTopicWidth, primaryTopic,
			snippet,
		)
	}

	return nil
}

// primaryTopicOf returns the first topic (always in normalized form), or an
// empty string if the entry has no topics.
func primaryTopicOf(entry storage.Entry) string {
	if len(entry.Topics) == 0 {
		return ""
	}
	return entry.Topics[0]
}
