package commands

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/skorokithakis/gnosis/internal/storage"
)

const (
	// staleEntryAge is how long an entry may go without being verified before
	// `gn help review` suggests checking it again. It is deliberately not
	// configurable: the threshold is a nudge to re-read an entry, not a policy,
	// and exposing a knob would add surface area without changing what the
	// block is for.
	staleEntryAge = 180 * 24 * time.Hour

	// housekeepingEntryLimit caps how many stale entries are listed. A review
	// should send the agent to inspect a handful of entries, not to re-read the
	// whole knowledge base; the oldest entries go first so repeated sessions
	// work through the backlog without newer-but-still-old entries crowding it
	// out.
	housekeepingEntryLimit = 5
)

// selectStaleEntries returns up to housekeepingEntryLimit entries that were last
// updated longer ago than staleEntryAge, oldest first. Ties are broken by ID so
// the output is deterministic. now is passed in rather than read from the clock
// so the selection can be tested at the threshold boundary.
func selectStaleEntries(entries []storage.Entry, now time.Time) []storage.Entry {
	var stale []storage.Entry
	for _, entry := range entries {
		if now.Sub(entry.UpdatedAt) > staleEntryAge {
			stale = append(stale, entry)
		}
	}

	sort.Slice(stale, func(first, second int) bool {
		if stale[first].UpdatedAt.Equal(stale[second].UpdatedAt) {
			return stale[first].ID < stale[second].ID
		}
		return stale[first].UpdatedAt.Before(stale[second].UpdatedAt)
	})

	if len(stale) > housekeepingEntryLimit {
		stale = stale[:housekeepingEntryLimit]
	}
	return stale
}

// Housekeeping appends the stale-entry reminder that `gn help review` prints
// after the command reference. The caller decides whether the repository has a
// knowledge base at all; when there are no entries, or none are stale, nothing
// is written.
func Housekeeping(store *storage.Store, now time.Time, writer io.Writer) error {
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	stale := selectStaleEntries(entries, now)
	if len(stale) == 0 {
		return nil
	}

	// Unique prefixes are computed against the whole store, not just the listed
	// entries, so a displayed prefix is guaranteed to resolve if copied into
	// `gn show`.
	allIDs := make([]string, len(entries))
	for index, entry := range entries {
		allIDs[index] = entry.ID
	}
	topicWidth := maxTopicWidth(stale)

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "HOUSEKEEPING")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "After you record this session's entries, check these entries. Nobody")
	fmt.Fprintln(writer, "has updated them in over 180 days:")
	fmt.Fprintln(writer)
	for _, entry := range stale {
		fmt.Fprintf(writer, "  %s\n", formatEntryLine(entry, allIDs, topicWidth))
	}
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, `For each one, read it with "gn show <id>" and compare it with the`)
	fmt.Fprintln(writer, `code. If it is still accurate, run "gn verify <id>". If it is partly`)
	fmt.Fprintln(writer, `wrong, fix it with "gn edit <id>". If it is obsolete, run "gn rm <id>".`)

	return nil
}
