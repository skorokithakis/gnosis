package commands

import (
	"fmt"
	"io"
	"time"

	"github.com/skorokithakis/gnosis/internal/storage"
	"github.com/skorokithakis/gnosis/internal/termcolor"
)

// Verify implements "gn verify <id> [<id>...]". It marks each named entry as
// still accurate by setting UpdatedAt to now, leaving its text, topics, author,
// and related IDs untouched. argv should be os.Args[2:] (everything after
// "verify"). The verified IDs are written to writer in the same unique-prefix
// form rm uses, so the user sees exactly which entries were touched.
func Verify(store *storage.Store, argv []string, writer io.Writer) error {
	if len(argv) == 0 {
		return fmt.Errorf("usage: gn verify <id> [<id>...]")
	}

	// Resolve every prefix before touching the file, matching rm: one mistyped
	// ID should leave the whole batch unverified rather than partially applied.
	// Update below re-reads under the exclusive lock, so entries appended
	// between this read and the rewrite are preserved.
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	allIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		allIDs = append(allIDs, entry.ID)
	}

	resolvedIDs := make([]string, len(argv))
	for index, prefix := range argv {
		resolvedID, err := ResolveIDPrefix(entries, prefix)
		if err != nil {
			return err
		}
		resolvedIDs[index] = resolvedID
	}

	verifySet := make(map[string]bool, len(resolvedIDs))
	for _, id := range resolvedIDs {
		verifySet[id] = true
	}

	// Update holds the exclusive lock across the read-modify-write cycle and
	// rewrites entries in ID order, so verification is safe alongside
	// concurrent writers and does not disturb the file's sort invariant.
	now := time.Now().UTC()
	err = store.Update(func(current []storage.Entry) []storage.Entry {
		for index, entry := range current {
			if verifySet[entry.ID] {
				updated := entry
				updated.UpdatedAt = now
				current[index] = updated
			}
		}
		return current
	})
	if err != nil {
		return fmt.Errorf("rewriting entries: %w", err)
	}

	for _, id := range resolvedIDs {
		fmt.Fprintln(writer, termcolor.UniqueID(id, allIDs))
	}

	return nil
}
