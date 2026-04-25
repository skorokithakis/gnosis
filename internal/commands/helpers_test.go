package commands_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// newTestStore creates a Store backed by a temporary directory.
func newTestStore(t *testing.T) *storage.Store {
	t.Helper()
	store, err := storage.NewStoreAt(filepath.Join(t.TempDir(), ".gnosis"), t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	return store
}

// sampleEntry returns a minimal valid Entry for use in tests.
func sampleEntry(id string) storage.Entry {
	now := time.Now().UTC().Truncate(time.Second)
	return storage.Entry{
		ID:        id,
		Topics:    []string{"GoLang", "Testing"},
		Text:      "some text",
		Related:   []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// appendEntries appends each entry to store, failing the test on error.
func appendEntries(t *testing.T, store *storage.Store, entries ...storage.Entry) {
	t.Helper()
	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append(%s): %v", entry.ID, err)
		}
	}
}
