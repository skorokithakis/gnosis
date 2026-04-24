package commands

import (
	"fmt"
	"io"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// Remove implements the "gnosis rm <id> [<id>...]" command. argv should be
// os.Args[2:] (everything after "rm"). stderr is used for warnings about
// dangling Related references; stdout receives the removed IDs.
func Remove(store *storage.Store, argv []string, stdout io.Writer, stderr io.Writer) error {
	if len(argv) == 0 {
		return fmt.Errorf("usage: gnosis rm <id> [<id>...]")
	}

	// Read entries once before the lock to validate the requested IDs. We
	// refuse the whole operation if any ID is unknown, so we need to check
	// before acquiring the exclusive lock. The locked Update below re-reads
	// entries, so any entries appended between here and the lock are preserved.
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	existingIDs := make(map[string]bool, len(entries))
	for _, entry := range entries {
		existingIDs[entry.ID] = true
	}

	// Validate all requested IDs before touching the file. Partial removal is
	// confusing, so we refuse the whole operation if any ID is unknown.
	removeSet := make(map[string]bool, len(argv))
	for _, id := range argv {
		if !existingIDs[id] {
			return fmt.Errorf("entry %q does not exist", id)
		}
		removeSet[id] = true
	}

	// warnings collects dangling-reference messages discovered inside the
	// transform. They are printed after Update returns so that no I/O happens
	// while the exclusive lock is held.
	var warnings []string

	err = store.Update(func(current []storage.Entry) []storage.Entry {
		warnings = warnings[:0]

		var surviving []storage.Entry
		for _, entry := range current {
			if !removeSet[entry.ID] {
				surviving = append(surviving, entry)
			}
		}

		// Warn about dangling Related references in surviving entries. We do
		// not modify those entries automatically because silent data mutation
		// would be surprising and hard to audit.
		for _, entry := range surviving {
			for _, relatedID := range entry.Related {
				if removeSet[relatedID] {
					warnings = append(warnings, fmt.Sprintf("note: entry %s references removed entry %s", entry.ID, relatedID))
				}
			}
		}

		return surviving
	})
	if err != nil {
		return fmt.Errorf("rewriting entries: %w", err)
	}

	for _, warning := range warnings {
		fmt.Fprintln(stderr, warning)
	}

	for _, id := range argv {
		fmt.Fprintln(stdout, id)
	}

	return nil
}
