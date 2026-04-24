package commands

import (
	"fmt"
	"io"
	"sort"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// TopicAggregate holds the normalized topic name and its entry count.
type TopicAggregate struct {
	Topic string
	Count int
}

// AggregateTopics scans all entries and returns one aggregate per topic.
// Because topics are always stored in normalized form, no further normalization
// is needed here — counting is a simple map lookup.
func AggregateTopics(entries []storage.Entry) []TopicAggregate {
	counts := map[string]int{}
	for _, entry := range entries {
		for _, topic := range entry.Topics {
			counts[topic]++
		}
	}

	aggregates := make([]TopicAggregate, 0, len(counts))
	for topic, count := range counts {
		aggregates = append(aggregates, TopicAggregate{Topic: topic, Count: count})
	}
	return aggregates
}

// sortTopics sorts aggregates in-place: descending by count, then ascending
// alphabetically by topic for ties.
func sortTopics(aggregates []TopicAggregate) {
	sort.Slice(aggregates, func(i, j int) bool {
		if aggregates[i].Count != aggregates[j].Count {
			return aggregates[i].Count > aggregates[j].Count
		}
		return aggregates[i].Topic < aggregates[j].Topic
	})
}

// Topics loads all entries from store, aggregates them by topic, and writes the
// sorted result to writer. If there are no entries, nothing is written.
func Topics(store *storage.Store, writer io.Writer) error {
	entries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	aggregates := AggregateTopics(entries)
	if len(aggregates) == 0 {
		return nil
	}

	sortTopics(aggregates)

	// Find the widest count so all counts are right-aligned to the same width.
	maxCount := aggregates[0].Count
	width := countWidth(maxCount)

	for _, aggregate := range aggregates {
		fmt.Fprintf(writer, "%*d  %s\n", width, aggregate.Count, aggregate.Topic)
	}

	return nil
}

// countWidth returns the number of decimal digits in n.
func countWidth(n int) int {
	if n == 0 {
		return 1
	}
	width := 0
	for n > 0 {
		width++
		n /= 10
	}
	return width
}
