package commands

import (
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// makeStaleTestEntry builds an entry for selection tests.
func makeStaleTestEntry(id string, updatedAt time.Time) storage.Entry {
	return storage.Entry{
		ID:        id,
		Topics:    []string{"golang"},
		Text:      "some text",
		Related:   []string{},
		CreatedAt: updatedAt,
		UpdatedAt: updatedAt,
	}
}

// TestSelectStaleEntries pins the selection contract in one fixture: the
// strictly-more-than-180-days threshold, oldest-first ordering, the five-entry
// limit, and the ID tie-break. Input order is scrambled, the oldest entry comes
// after the first five, and the two entries tied on updated_at are given in
// reverse ID order, so an implementation that truncates before sorting or falls
// back to input order fails. A recently verified entry has an old CreatedAt to
// catch selection by the wrong timestamp.
func TestSelectStaleEntries(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour

	tiedUpdatedAt := now.Add(-200 * day)
	recentlyVerified := makeStaleTestEntry("jjjjjj", now.Add(-time.Hour))
	recentlyVerified.CreatedAt = now.Add(-3 * 365 * day)

	entries := []storage.Entry{
		makeStaleTestEntry("bbbbbb", now.Add(-250*day)),
		makeStaleTestEntry("dddddd", tiedUpdatedAt),
		makeStaleTestEntry("cccccc", tiedUpdatedAt),
		makeStaleTestEntry("eeeeee", now.Add(-staleEntryAge-time.Second)),
		makeStaleTestEntry("ffffff", now.Add(-staleEntryAge-500*time.Millisecond)),
		makeStaleTestEntry("aaaaaa", now.Add(-300*day)),
		makeStaleTestEntry("hhhhhh", now.Add(-staleEntryAge)),
		recentlyVerified,
	}

	stale := selectStaleEntries(entries, now)

	want := []string{"aaaaaa", "bbbbbb", "cccccc", "dddddd", "eeeeee"}
	if len(stale) != len(want) {
		t.Fatalf("expected %d stale entries, got %d: %v", len(want), len(stale), stale)
	}
	for index, wantID := range want {
		if stale[index].ID != wantID {
			t.Errorf("stale[%d] = %s, want %s; full result: %v", index, stale[index].ID, wantID, stale)
		}
	}

	// With more than five stale entries the limit would hide the exact-boundary
	// entry even if it were wrongly treated as stale, so isolate the boundary in
	// a second small fixture. This is what makes the "strictly more than" rule
	// observable rather than masked by truncation.
	boundary := []storage.Entry{
		makeStaleTestEntry("hhhhhh", now.Add(-staleEntryAge)),
		makeStaleTestEntry("eeeeee", now.Add(-staleEntryAge-time.Second)),
	}
	stale = selectStaleEntries(boundary, now)
	if len(stale) != 1 || stale[0].ID != "eeeeee" {
		t.Errorf("expected only the entry just over the boundary, got %v", stale)
	}
}
