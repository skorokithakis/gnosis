package commands

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/storage"
	"github.com/skorokithakis/gnosis/internal/termcolor"
)

// nonAlphanumericPattern matches one or more consecutive characters that are
// neither alphanumeric nor whitespace. These are replaced with a single space
// when sanitizing a bareword so that e.g. "keymaster-token-auth" becomes
// "keymaster token auth", which FTS5 interprets as an AND-of-three-terms match.
var nonAlphanumericPattern = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

// whitespaceRunPattern matches one or more consecutive whitespace characters,
// including newlines. Used to collapse multi-line FTS5 snippets to a single line.
var whitespaceRunPattern = regexp.MustCompile(`\s+`)

// sanitizeQuery prepares a user-supplied FTS5 query string by scanning
// left-to-right and applying four rules:
//  1. Quoted phrases "..." pass through verbatim (including the quotes).
//  2. Structural chars (, ), :, and whitespace pass through verbatim.
//  3. Whole-word uppercase tokens AND, OR, NOT pass through verbatim.
//  4. Every other bareword run has non-alphanumeric chars replaced with spaces.
//
// This allows queries like "foo OR latest-release" to work: the OR operator is
// preserved while the hyphen in "latest-release" is sanitized so that FTS5 does
// not parse it as "latest - release" (column-filter operator).
func sanitizeQuery(query string) string {
	var result strings.Builder
	i := 0
	n := len(query)

	for i < n {
		c := query[i]

		// Rule 1: Quoted phrases "..." pass through verbatim.
		if c == '"' {
			// Scan to the closing quote.
			j := i + 1
			for j < n && query[j] != '"' {
				j++
			}
			if j < n {
				j++ // include closing quote
			}
			result.WriteString(query[i:j])
			i = j
			continue
		}

		// Rule 2: Structural chars and whitespace pass through verbatim.
		if c == '(' || c == ')' || c == ':' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			result.WriteByte(c)
			i++
			continue
		}

		// Collect a bareword run: everything until a structural char,
		// whitespace, or quote.
		j := i
		for j < n {
			cc := query[j]
			if cc == '"' || cc == '(' || cc == ')' || cc == ':' ||
				cc == ' ' || cc == '\t' || cc == '\n' || cc == '\r' {
				break
			}
			j++
		}

		word := query[i:j]

		// Rule 3: Whole-word uppercase AND, OR, NOT pass through verbatim.
		if word == "AND" || word == "OR" || word == "NOT" {
			result.WriteString(word)
		} else {
			// Rule 4: Sanitize the bareword.
			result.WriteString(nonAlphanumericPattern.ReplaceAllString(word, " "))
		}

		i = j
	}

	return result.String()
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
		return fmt.Errorf("usage: gn search <query> [--limit N]")
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
	// for each hit without a linear scan per result. Also collect all IDs so
	// that UniqueID can determine the shortest unambiguous prefix for each hit.
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}
	entryByID := make(map[string]storage.Entry, len(entries))
	allIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		entryByID[entry.ID] = entry
		allIDs = append(allIDs, entry.ID)
	}

	// Determine the maximum topic width across all hits so that snippets start
	// at a consistent horizontal position regardless of topic name length.
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
		updatedDate := strings.Repeat(" ", 10)
		if found {
			primaryTopic = primaryTopicOf(entry)
			updatedDate = entry.UpdatedAt.Format("2006-01-02")
		}
		// Collapse all whitespace runs (including newlines that FTS5's
		// snippet() may include from the indexed text) to single spaces so
		// that each result occupies exactly one output line.
		snippet := whitespaceRunPattern.ReplaceAllString(hit.Snippet, " ")
		// Render match-delimiter sentinels emitted by FTS5's snippet() as
		// bold red on color terminals, falling back to literal brackets
		// elsewhere.
		snippet = termcolor.HighlightMatches(snippet, index.MatchStart, index.MatchEnd)

		// IDs are always 6 bytes, so no id-side padding is needed.
		// We color the id, date, and topic separately and then append explicit
		// spaces for the topic column, because fmt.Fprintf's %-*s measures
		// width by byte length — ANSI escape sequences would inflate the count
		// and misalign the snippet column.
		coloredID := termcolor.UniqueID(hit.EntryID, allIDs)
		coloredDate := termcolor.Date(updatedDate)
		coloredTopic := termcolor.Topic(primaryTopic)
		topicPad := strings.Repeat(" ", maxTopicWidth-len(primaryTopic))
		fmt.Fprintf(writer, "%s  %s  %s%s  %s\n", coloredID, coloredDate, coloredTopic, topicPad, snippet)
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
