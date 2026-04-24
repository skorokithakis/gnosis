package commands

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/skorokithakis/gnosis/internal/storage"
	"github.com/skorokithakis/gnosis/internal/termcolor"
	"github.com/skorokithakis/gnosis/internal/textwrap"
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
//
// wrapWidth controls body text wrapping via textwrap.Wrap. A value of zero or
// less disables wrapping, which is appropriate for non-TTY output.
func Show(store *storage.Store, target string, wrapWidth int, writer io.Writer) error {
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	allIDs := make([]string, len(entries))
	for index, entry := range entries {
		allIDs[index] = entry.ID
	}

	if utf8.RuneCountInString(target) <= storage.IDLength {
		resolvedID, err := ResolveIDPrefix(entries, target)
		if err != nil {
			return err
		}
		return showByID(entries, resolvedID, allIDs, wrapWidth, writer)
	}
	return showByTopic(entries, target, allIDs, wrapWidth, writer)
}

// showByID finds the single entry whose ID matches target and prints its full
// details. It returns an error if no entry with that ID exists.
func showByID(entries []storage.Entry, target string, allIDs []string, wrapWidth int, writer io.Writer) error {
	for _, entry := range entries {
		if entry.ID == target {
			printEntry(entry, allIDs, wrapWidth, writer)
			return nil
		}
	}
	return fmt.Errorf("entry %q not found", target)
}

// showByTopic normalises target, finds all entries that carry that topic, sorts
// them by creation time, and prints a header followed by each entry.
func showByTopic(entries []storage.Entry, target string, allIDs []string, wrapWidth int, writer io.Writer) error {
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
	fmt.Fprintf(writer, "%s %s  (%s %s)\n\n",
		termcolor.Dim("topic:"),
		termcolor.Topic(normalizedTarget),
		termcolor.Bold(fmt.Sprintf("%d", len(matched))),
		entryWord,
	)

	for index, entry := range matched {
		printEntry(entry, allIDs, wrapWidth, writer)
		if index < len(matched)-1 {
			fmt.Fprintln(writer)
		}
	}

	return nil
}

// printEntry writes a single entry to writer in the human-readable terminal
// format. The layout is:
//
//	id:      <colored unique-prefix id>
//	topics:  <cyan topic>, <cyan topic>
//	related: <colored id>, <colored id>   (omitted when entry.Related is empty)
//	created: <yellow date>
//	updated: <yellow date>
//
//	<body, wrapped via textwrap.Wrap when wrapWidth > 0>
//
// Labels are rendered dim/faint. allIDs is the full set of IDs in the store,
// forwarded to termcolor.UniqueID so it can determine the shortest unique
// prefix for each ID.
func printEntry(entry storage.Entry, allIDs []string, wrapWidth int, writer io.Writer) {
	coloredTopics := make([]string, len(entry.Topics))
	for index, topic := range entry.Topics {
		coloredTopics[index] = termcolor.Topic(topic)
	}

	createdDate := entry.CreatedAt.Format("2006-01-02")
	updatedDate := entry.UpdatedAt.Format("2006-01-02")

	fmt.Fprintf(writer, "%s %s\n", termcolor.Dim("id:"), termcolor.UniqueID(entry.ID, allIDs))
	fmt.Fprintf(writer, "%s %s\n", termcolor.Dim("topics:"), strings.Join(coloredTopics, ", "))

	if len(entry.Related) > 0 {
		coloredRelated := make([]string, len(entry.Related))
		for index, relatedID := range entry.Related {
			coloredRelated[index] = termcolor.UniqueID(relatedID, allIDs)
		}
		fmt.Fprintf(writer, "%s %s\n", termcolor.Dim("related:"), strings.Join(coloredRelated, ", "))
	}

	fmt.Fprintf(writer, "%s %s\n", termcolor.Dim("created:"), termcolor.Date(createdDate))
	fmt.Fprintf(writer, "%s %s\n", termcolor.Dim("updated:"), termcolor.Date(updatedDate))

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, textwrap.Wrap(entry.Text, wrapWidth))
}
