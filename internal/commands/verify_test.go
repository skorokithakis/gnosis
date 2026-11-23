package commands_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/commands"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// findEntry returns the entry with id, failing the test if it is absent.
func findEntry(t *testing.T, store *storage.Store, id string) storage.Entry {
	t.Helper()
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	for _, entry := range entries {
		if entry.ID == id {
			return entry
		}
	}
	t.Fatalf("entry %s not found", id)
	return storage.Entry{}
}

// TestVerify_updates_only_updated_at_and_accepts_multiple_ids verifies the core
// contract (only UpdatedAt may change) and that a batch can mix a prefix and a
// full ID, updating those entries while leaving the rest of the store alone.
func TestVerify_updates_only_updated_at_and_accepts_multiple_ids(t *testing.T) {
	store := newTestStore(t)

	createdAt := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	oldUpdatedAt := time.Date(2021, 6, 7, 8, 9, 10, 0, time.UTC)
	original := storage.Entry{
		ID:        "aaaaaa",
		Topics:    []string{"golang", "testing"},
		Text:      "some text\nwith two lines",
		Author:    "A Person <person@example.com>",
		Related:   []string{"bbbbbb"},
		CreatedAt: createdAt,
		UpdatedAt: oldUpdatedAt,
	}
	second := sampleEntry("bbbbbb")
	second.UpdatedAt = oldUpdatedAt
	third := sampleEntry("cccccc")
	third.UpdatedAt = oldUpdatedAt
	appendEntries(t, store, original, second, third)

	var output strings.Builder
	if err := commands.Verify(store, []string{"aaa", "bbbbbb"}, &output); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	verified := findEntry(t, store, "aaaaaa")
	if !verified.UpdatedAt.After(oldUpdatedAt) {
		t.Errorf("UpdatedAt = %v, want a time after %v", verified.UpdatedAt, oldUpdatedAt)
	}
	if !verified.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt changed: %v, want %v", verified.CreatedAt, createdAt)
	}
	if !reflect.DeepEqual(verified.Topics, original.Topics) {
		t.Errorf("Topics changed: %v, want %v", verified.Topics, original.Topics)
	}
	if verified.Text != original.Text {
		t.Errorf("Text changed: %q, want %q", verified.Text, original.Text)
	}
	if verified.Author != original.Author {
		t.Errorf("Author changed: %q, want %q", verified.Author, original.Author)
	}
	if !reflect.DeepEqual(verified.Related, original.Related) {
		t.Errorf("Related changed: %v, want %v", verified.Related, original.Related)
	}

	if !findEntry(t, store, "bbbbbb").UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected bbbbbb to be updated by the full-ID argument")
	}
	if !findEntry(t, store, "cccccc").UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("expected cccccc to remain untouched")
	}

	for _, id := range []string{"aaaaaa", "bbbbbb"} {
		if !strings.Contains(output.String(), id) {
			t.Errorf("expected verified ID %s on stdout, got %q", id, output.String())
		}
	}
}

// TestVerify_invalid_batch_writes_nothing covers the batch semantics shared
// with rm: one unknown or ambiguous prefix must abort the whole invocation
// before any entry is touched, even when an earlier argument was valid.
func TestVerify_invalid_batch_writes_nothing(t *testing.T) {
	testCases := []struct {
		name          string
		argv          []string
		wantErrorPart string
	}{
		{
			name:          "unknown prefix",
			argv:          []string{"aaaaaa", "zzzzzz"},
			wantErrorPart: `entry "zzzzzz" not found`,
		},
		{
			name:          "ambiguous prefix",
			argv:          []string{"aaaaaa", "aaaaa"},
			wantErrorPart: "ambiguous",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := newTestStore(t)
			oldUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			first := sampleEntry("aaaaaa")
			first.UpdatedAt = oldUpdatedAt
			second := sampleEntry("aaaaab")
			second.UpdatedAt = oldUpdatedAt
			appendEntries(t, store, first, second)

			var output strings.Builder
			err := commands.Verify(store, testCase.argv, &output)
			if err == nil {
				t.Fatal("expected error for invalid ID, got nil")
			}
			if !strings.Contains(err.Error(), testCase.wantErrorPart) {
				t.Errorf("error %q does not contain %q", err.Error(), testCase.wantErrorPart)
			}
			for _, id := range []string{"aaaaaa", "aaaaab"} {
				if !findEntry(t, store, id).UpdatedAt.Equal(oldUpdatedAt) {
					t.Errorf("entry %s was modified despite the invalid batch", id)
				}
			}
			if output.Len() != 0 {
				t.Errorf("expected no output on error, got %q", output.String())
			}
		})
	}
}
