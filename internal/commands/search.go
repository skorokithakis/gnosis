package commands

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/storage"
	"github.com/skorokithakis/gnosis/internal/termcolor"
)

// nonAlphanumericPattern matches one or more consecutive characters that are
// neither a Unicode letter or digit nor whitespace. These are replaced with a
// single space when sanitizing a bareword so that e.g. "keymaster-token-auth"
// becomes "keymaster token auth", which FTS5 interprets as an AND-of-three-terms
// match. Unicode classes are used so that accented letters ("café") and CJK
// characters survive sanitization instead of being turned into spaces.
var nonAlphanumericPattern = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)

// whitespaceRunPattern matches one or more consecutive whitespace characters,
// including newlines. Used to collapse multi-line FTS5 snippets to a single line.
var whitespaceRunPattern = regexp.MustCompile(`\s+`)

// sanitizeQuery prepares a user-supplied FTS5 query string by scanning
// left-to-right and applying four rules:
//  1. Quoted phrases "..." pass through with their content run through
//     index.SplitCJK, so '"非法状态"' becomes '"非 法 状 态"'. Quoted content
//     without CJK is left byte-for-byte unchanged.
//  2. Structural chars (, ), :, and whitespace pass through verbatim.
//  3. Whole-word uppercase tokens AND, OR, NOT pass through verbatim.
//  4. Every other bareword run keeps Unicode letters and digits, replaces the
//     rest with spaces, and turns each maximal run of CJK characters into a
//     quoted phrase of single characters: 'API状态' becomes 'API "状 态"'.
//
// This allows queries like "foo OR latest-release" to work: the OR operator is
// preserved while the hyphen in "latest-release" is sanitized so that FTS5 does
// not parse it as "latest - release" (column-filter operator). The CJK phrases
// are needed because the index stores each CJK character as its own token, so a
// bare CJK run would otherwise be looked up as one unmatchable token.
func sanitizeQuery(query string) string {
	var result strings.Builder
	i := 0
	n := len(query)

	for i < n {
		c := query[i]

		// Rule 1: Quoted phrases "..." pass through, with their content split
		// the same way the index splits it.
		if c == '"' {
			// Scan to the closing quote.
			j := i + 1
			for j < n && query[j] != '"' {
				j++
			}
			contentEnd := j
			if j < n {
				j++ // include closing quote
			}
			result.WriteByte('"')
			result.WriteString(index.SplitCJK(query[i+1 : contentEnd]))
			if contentEnd < j {
				result.WriteByte('"')
			}
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
			// Rule 4: Sanitize the bareword, then phrase its CJK runs.
			sanitized := nonAlphanumericPattern.ReplaceAllString(word, " ")
			result.WriteString(quoteCJKRuns(sanitized))
		}

		i = j
	}

	return result.String()
}

// quoteCJKRuns replaces every maximal run of CJK characters in text with a
// quoted FTS5 phrase of space-separated single characters. A space is written
// between a quoted phrase and neighbouring text so FTS5 sees separate terms;
// existing spaces are not doubled and no trailing space is left behind.
func quoteCJKRuns(text string) string {
	runes := []rune(text)

	var builder strings.Builder
	builder.Grow(len(text) + len(text)/2)
	// spacePending is set after a quoted phrase so the following rune is
	// separated from the closing quote. lastWasSpace tracks whether a separator
	// already precedes the current position.
	spacePending := false
	lastWasSpace := true // the start of the text behaves like a separator

	for i := 0; i < len(runes); {
		r := runes[i]

		if index.IsCJK(r) {
			if !lastWasSpace {
				builder.WriteByte(' ')
			}
			builder.WriteByte('"')
			for {
				builder.WriteRune(runes[i])
				i++
				if i >= len(runes) || !index.IsCJK(runes[i]) {
					break
				}
				builder.WriteByte(' ')
			}
			builder.WriteByte('"')
			spacePending = true
			lastWasSpace = false
			continue
		}

		if r == ' ' {
			builder.WriteByte(' ')
			spacePending = false
			lastWasSpace = true
			i++
			continue
		}

		if spacePending {
			builder.WriteByte(' ')
			spacePending = false
		}
		builder.WriteRune(r)
		lastWasSpace = false
		i++
	}

	return builder.String()
}

// hasSearchableTerms reports whether query contains at least one Unicode letter
// or digit. A query with none (for example "---" or an empty quoted phrase)
// would only reach FTS5 as operators and punctuation, where it fails with a
// syntax error, so Search rejects it before touching the index.
func hasSearchableTerms(query string) bool {
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
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

	if !hasSearchableTerms(query) {
		return fmt.Errorf("query has no searchable terms")
	}

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
