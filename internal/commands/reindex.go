package commands

import (
	"fmt"
	"io"

	"github.com/skorokithakis/gnosis/internal/index"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// Reindex implements the `gnosis reindex` command. It forces a full rebuild of
// the SQLite FTS5 index from the JSONL source of truth, regardless of whether
// the index is already fresh. This is useful after manual JSONL edits (e.g.
// git merges) or when the cache has become inconsistent.
func Reindex(repoRoot string, store *storage.Store, writer io.Writer) error {
	idx, err := index.Open(repoRoot, store)
	if err != nil {
		return fmt.Errorf("opening index: %w", err)
	}
	defer idx.Close()

	if err := idx.Rebuild(); err != nil {
		return fmt.Errorf("rebuilding index: %w", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	entryWord := "entries"
	if len(entries) == 1 {
		entryWord = "entry"
	}
	fmt.Fprintf(writer, "Rebuilt index: %d %s indexed at %s\n", len(entries), entryWord, idx.DBPath())
	return nil
}
