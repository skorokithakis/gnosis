package commands

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// Show implements the `gnosis show <target>` command. It writes output to
// writer so that callers (including tests) can capture it without redirecting
// os.Stdout. The returned error is non-nil only for I/O or storage failures;
// "not found" conditions are reported by returning an error with a descriptive
// message so the caller can exit with status 1.
//
// Dispatch is purely by query length: a target of 6 characters or fewer is
// treated as an ID prefix; 7 characters or more is treated as a topic name.
// This removes the ambiguity of the old ID→topic fallback while still letting
// short English words work as topic names when they are long enough (≥7 chars).
func Show(store *storage.Store, target string, writer io.Writer) error {
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	if utf8.RuneCountInString(target) <= storage.IDLength {
		resolvedID, err := ResolveIDPrefix(entries, target)
		if err != nil {
			return err
		}
		return showByID(entries, resolvedID, writer)
	}
	return showByTopic(entries, target, writer)
}

// showByID finds the single entry whose ID matches target and prints its full
// details. It returns an error if no entry with that ID exists.
func showByID(entries []storage.Entry, target string, writer io.Writer) error {
	for _, entry := range entries {
		if entry.ID == target {
			printEntry(entry, writer)
			return nil
		}
	}
	return fmt.Errorf("entry %q not found", target)
}

// showByTopic normalises target, finds all entries that carry that topic, sorts
// them by creation time, and prints a header followed by each entry.
func showByTopic(entries []storage.Entry, target string, writer io.Writer) error {
	normalizedTarget := storage.NormalizeTopic(target)

	var matched []storage.Entry
	for _, entry := range entries {
		for _, topic := range entry.Topics {
			if topic == normalizedTarget {
				matched = append(matched, entry)
				break
			}
		}
	}

	if len(matched) == 0 {
		return fmt.Errorf("no entries found for topic %q", target)
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.Before(matched[j].CreatedAt)
	})

	// Topics are always stored normalized, so the header uses the normalized
	// form directly rather than trying to recover the original display form.
	entryWord := "entries"
	if len(matched) == 1 {
		entryWord = "entry"
	}
	fmt.Fprintf(writer, "Topic: %s  (%d %s)\n\n", normalizedTarget, len(matched), entryWord)

	for index, entry := range matched {
		printEntry(entry, writer)
		if index < len(matched)-1 {
			fmt.Fprintln(writer)
		}
	}

	return nil
}

// printEntry writes a single entry to writer in the human-readable terminal
// format. The layout is:
//
//	<id>  [Topic1, Topic2]  created <date>  updated <date>
//	Related: <id1>, <id2>
//
//	<text body>
func printEntry(entry storage.Entry, writer io.Writer) {
	topicsDisplay := "[" + strings.Join(entry.Topics, ", ") + "]"
	createdDate := entry.CreatedAt.Format("2006-01-02")
	updatedDate := entry.UpdatedAt.Format("2006-01-02")

	fmt.Fprintf(writer, "%s  %s  created %s  updated %s\n",
		entry.ID, topicsDisplay, createdDate, updatedDate)

	if len(entry.Related) > 0 {
		fmt.Fprintf(writer, "Related: %s\n", strings.Join(entry.Related, ", "))
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, entry.Text)
}
