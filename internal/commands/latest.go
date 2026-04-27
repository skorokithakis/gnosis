package commands

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/skorokithakis/gnosis/internal/storage"
	"github.com/skorokithakis/gnosis/internal/termcolor"
)

const defaultLatestLimit = 20

// snippetMaxRunes is the maximum number of Unicode code points shown in a
// latest snippet. Entries can be arbitrarily long, and without an FTS5 query
// there is no natural excerpt to anchor on, so we cap at a fixed width to keep
// output scannable. The ellipsis appended when text is truncated is not counted
// against this limit.
const snippetMaxRunes = 200

// Latest implements the "gn latest [--limit N]" command. It reads all entries,
// sorts them newest-first by CreatedAt, and prints up to limit entries in the
// same columnar format as Search (id, primary topic, text snippet). argv should
// be os.Args[2:] (everything after "latest"). Output is written to writer so
// callers and tests can capture it.
func Latest(store *storage.Store, argv []string, writer io.Writer) error {
	limit := defaultLatestLimit

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
			return fmt.Errorf("usage: gn latest [--limit N]")
		}
	}

	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	// Sort a copy so we do not mutate the slice returned by ReadAll. Newest
	// entries first; ties broken by ID for deterministic output.
	sorted := make([]storage.Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CreatedAt.Equal(sorted[j].CreatedAt) {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	if limit < len(sorted) {
		sorted = sorted[:limit]
	}

	// Collect all IDs from the full entry set so UniqueID can determine the
	// shortest unambiguous prefix across the whole store, not just the visible
	// window.
	allIDs := make([]string, len(entries))
	for index, entry := range entries {
		allIDs[index] = entry.ID
	}

	// Determine the maximum primary-topic width across the visible results so
	// that snippets start at a consistent horizontal position.
	maxTopicWidth := 0
	for _, entry := range sorted {
		topic := primaryTopicOf(entry)
		if len(topic) > maxTopicWidth {
			maxTopicWidth = len(topic)
		}
	}

	for _, entry := range sorted {
		primaryTopic := primaryTopicOf(entry)
		// Collapse whitespace runs (including embedded newlines) to single
		// spaces so each result occupies exactly one output line.
		snippet := whitespaceRunPattern.ReplaceAllString(entry.Text, " ")
		// Truncate by rune count, not byte count, so multi-byte characters
		// do not produce a partial code point at the cut point.
		if runes := []rune(snippet); len(runes) > snippetMaxRunes {
			snippet = string(runes[:snippetMaxRunes]) + "…"
		}

		coloredID := termcolor.UniqueID(entry.ID, allIDs)
		coloredTopic := termcolor.Topic(primaryTopic)
		topicPad := strings.Repeat(" ", maxTopicWidth-len(primaryTopic))
		fmt.Fprintf(writer, "%s  %s%s  %s\n", coloredID, coloredTopic, topicPad, snippet)
	}

	return nil
}
