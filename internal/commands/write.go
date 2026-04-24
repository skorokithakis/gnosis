package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// Write implements the "gnosis write <topics> <text> [--related id,id]" command.
// argv should be os.Args[2:] (everything after "write"). The new entry ID is
// written to writer, matching the io.Writer pattern used by other commands.
func Write(store *storage.Store, argv []string, writer io.Writer) error {
	if len(argv) == 0 {
		return fmt.Errorf("usage: gnosis write <topics> <text> [--related id,id]")
	}

	topicsArg := argv[0]
	argv = argv[1:]

	// Parse --related before consuming the text argument, so that the flag can
	// appear anywhere after the topics argument.
	var relatedArg string
	var remaining []string
	for index := 0; index < len(argv); index++ {
		if argv[index] == "--related" {
			if index+1 >= len(argv) {
				return fmt.Errorf("--related requires a value")
			}
			relatedArg = argv[index+1]
			index++
		} else {
			remaining = append(remaining, argv[index])
		}
	}

	// The text is the first non-flag argument. If absent and stdin is not a
	// TTY, read from stdin so that piped input works (e.g. echo ... | gnosis write foo).
	var text string
	if len(remaining) > 0 {
		text = remaining[0]
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		text = string(raw)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("text must not be empty")
	}

	// Parse and deduplicate topics by normalized form. Storage will normalize
	// on write, but we deduplicate here so that "Foo,foo" within a single
	// write call collapses to one entry rather than two identical normalized
	// topics.
	rawTopics := strings.Split(topicsArg, ",")
	var topics []string
	seenNormalized := map[string]bool{}
	for _, raw := range rawTopics {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return fmt.Errorf("topic must not be empty")
		}
		normalized := storage.NormalizeTopic(trimmed)
		if normalized == "" {
			return fmt.Errorf("topic %q normalizes to empty string", trimmed)
		}
		if seenNormalized[normalized] {
			continue
		}
		seenNormalized[normalized] = true
		topics = append(topics, normalized)
	}

	// Load existing entries to validate --related IDs before attempting the
	// write. This read is outside the lock; AppendNew re-reads under the lock
	// to generate a collision-free ID atomically.
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	existingIDs := make(map[string]bool, len(entries))
	for _, entry := range entries {
		existingIDs[entry.ID] = true
	}

	// Validate --related IDs.
	var related []string
	if relatedArg != "" {
		for _, rawID := range strings.Split(relatedArg, ",") {
			id := strings.TrimSpace(rawID)
			if id == "" {
				return fmt.Errorf("related ID must not be empty")
			}
			if !existingIDs[id] {
				return fmt.Errorf("related ID %q does not exist", id)
			}
			related = append(related, id)
		}
	}

	now := time.Now().UTC()

	// AppendNew generates the ID atomically under a shared lock, preventing
	// two concurrent writers from independently reading the same ID set and
	// generating the same ID.
	newID, err := store.AppendNew(storage.Entry{
		Topics:    topics,
		Text:      text,
		Related:   related,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return fmt.Errorf("appending entry: %w", err)
	}

	fmt.Fprintln(writer, newID)
	return nil
}
