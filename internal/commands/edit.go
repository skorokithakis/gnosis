package commands

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// normalizeAndDeduplicateTopics normalizes each topic and removes duplicates
// that collapse to the same normalized form. It returns an error if any topic
// normalizes to empty (e.g. "---"), because storing an empty topic would
// corrupt the entry silently.
func normalizeAndDeduplicateTopics(raw []string) ([]string, error) {
	seen := map[string]bool{}
	var result []string
	for _, topic := range raw {
		normalized := storage.NormalizeTopic(topic)
		if normalized == "" {
			return nil, fmt.Errorf("topic %q normalizes to empty string", topic)
		}
		// Topics must be at least 7 characters in their normalized form so they
		// cannot be confused with 6-character entry ID prefixes during lookup.
		// Rune count is used so that multibyte characters are measured correctly.
		if utf8.RuneCountInString(normalized) < storage.IDLength+1 {
			return nil, fmt.Errorf("topic %q is too short (normalized form %q has %d characters, minimum is 7)", topic, normalized, utf8.RuneCountInString(normalized))
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	return result, nil
}

// ParsedEntry holds the parsed result of the temp file the user edited.
// It is exported so that tests in the commands_test package can inspect it.
type ParsedEntry struct {
	Topics  []string
	Related []string
	Text    string
}

// FormatEditBuffer serialises an entry into the editable temp-file format.
// The header lines carry topics and related IDs; a sentinel line separates
// them from the free-form body. This format is chosen so that '#' lines are
// clearly metadata and the separator is unambiguous even when the body
// contains '#' characters.
func FormatEditBuffer(entry storage.Entry) string {
	related := strings.Join(entry.Related, ", ")
	topics := strings.Join(entry.Topics, ", ")
	return fmt.Sprintf("# Topics: %s\n# Related: %s\n# ---\n%s", topics, related, entry.Text)
}

// ParseEditBuffer parses the content written by the user back into a
// ParsedEntry. Lines before the "# ---" separator that start with '#' are
// treated as header lines; everything after the separator is the body.
func ParseEditBuffer(content string) (ParsedEntry, error) {
	lines := strings.Split(content, "\n")

	var result ParsedEntry
	separatorIndex := -1

	for index, line := range lines {
		if line == "# ---" {
			separatorIndex = index
			break
		}
	}

	if separatorIndex == -1 {
		return ParsedEntry{}, fmt.Errorf("separator line '# ---' not found in edited file")
	}

	for _, line := range lines[:separatorIndex] {
		if !strings.HasPrefix(line, "#") {
			continue
		}
		// Strip the leading '#' and one optional space.
		trimmed := strings.TrimPrefix(line, "#")
		trimmed = strings.TrimPrefix(trimmed, " ")

		switch {
		case strings.HasPrefix(trimmed, "Topics:"):
			value := strings.TrimPrefix(trimmed, "Topics:")
			value = strings.TrimSpace(value)
			if value != "" {
				var rawTopics []string
				for _, topic := range strings.Split(value, ",") {
					topic = strings.TrimSpace(topic)
					if topic != "" {
						rawTopics = append(rawTopics, topic)
					}
				}
				normalized, err := normalizeAndDeduplicateTopics(rawTopics)
				if err != nil {
					return ParsedEntry{}, err
				}
				result.Topics = normalized
			}
		case strings.HasPrefix(trimmed, "Related:"):
			value := strings.TrimPrefix(trimmed, "Related:")
			value = strings.TrimSpace(value)
			if value != "" {
				for _, id := range strings.Split(value, ",") {
					id = strings.TrimSpace(id)
					if id != "" {
						result.Related = append(result.Related, id)
					}
				}
			}
		}
	}

	bodyLines := lines[separatorIndex+1:]
	result.Text = strings.TrimSpace(strings.Join(bodyLines, "\n"))

	return result, nil
}

// validateEditedEntry checks that the parsed edit result is internally
// consistent and that all referenced related IDs exist in the store (excluding
// the entry being edited, since self-references are meaningless). It resolves
// each related value via ResolveIDPrefix so that the caller can store full IDs
// regardless of whether the user typed a prefix or a complete ID. The returned
// slice contains the resolved full IDs in the same order as parsed.Related.
func validateEditedEntry(parsed ParsedEntry, entryID string, allEntries []storage.Entry) ([]string, error) {
	if len(parsed.Topics) == 0 {
		return nil, fmt.Errorf("topics must not be empty")
	}
	if parsed.Text == "" {
		return nil, fmt.Errorf("text body must not be empty")
	}

	// Build the candidate list excluding the entry being edited so that
	// self-references are rejected rather than silently accepted.
	var otherEntries []storage.Entry
	for _, entry := range allEntries {
		if entry.ID != entryID {
			otherEntries = append(otherEntries, entry)
		}
	}

	resolvedRelated := make([]string, 0, len(parsed.Related))
	for _, rawRelated := range parsed.Related {
		resolvedID, err := ResolveIDPrefix(otherEntries, rawRelated)
		if err != nil {
			return nil, fmt.Errorf("related ID %q: %w", rawRelated, err)
		}
		resolvedRelated = append(resolvedRelated, resolvedID)
	}

	return resolvedRelated, nil
}

// Edit loads the entry matching the given ID prefix, opens it in $EDITOR for
// the user to modify, then validates and saves the result. If the user makes no
// changes, it prints "no changes" and exits cleanly.
//
// The entry is read before the editor is opened so the user sees the current
// state. The actual save uses Store.Update, which holds an exclusive lock
// across read-modify-write, preventing concurrent appends from being lost.
func Edit(store *storage.Store, idPrefix string) error {
	// Read the entry before opening the editor so the user sees the current
	// state. This read is outside the lock because the editor interaction is
	// unbounded in time — we cannot hold a lock while waiting for the user.
	// The lock is acquired only during the final atomic rewrite.
	allEntries, err := store.ReadAll()
	if err != nil {
		return fmt.Errorf("reading entries: %w", err)
	}

	id, err := ResolveIDPrefix(allEntries, idPrefix)
	if err != nil {
		return err
	}

	entryIndex := -1
	for index, entry := range allEntries {
		if entry.ID == id {
			entryIndex = index
			break
		}
	}

	if entryIndex == -1 {
		return fmt.Errorf("entry %q not found", id)
	}

	original := allEntries[entryIndex]

	tmpFile, err := os.CreateTemp("", "gnosis-edit-*.txt")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(FormatEditBuffer(original)); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	editorEnv := os.Getenv("EDITOR")
	if editorEnv == "" {
		editorEnv = "vi"
	}

	// $EDITOR may contain flags (e.g. "code --wait", "vim -n"). We split on
	// whitespace and pass the first field as the program so that exec.Command
	// does not try to look up the entire string as a binary name. Invoking via
	// /bin/sh -c would work but is fragile with paths that contain spaces.
	editorFields := strings.Fields(editorEnv)
	editorArgs := append(editorFields[1:], tmpPath)
	cmd := exec.Command(editorFields[0], editorArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	parsed, err := ParseEditBuffer(string(editedContent))
	if err != nil {
		return err
	}

	// Treat nil and empty slice as equivalent when comparing related lists so
	// that a freshly-loaded entry with no related IDs (stored as null in JSON)
	// is not considered changed when the user leaves the Related line empty.
	originalRelated := original.Related
	if originalRelated == nil {
		originalRelated = []string{}
	}
	parsedRelated := parsed.Related
	if parsedRelated == nil {
		parsedRelated = []string{}
	}

	unchanged := slices.Equal(original.Topics, parsed.Topics) &&
		slices.Equal(originalRelated, parsedRelated) &&
		original.Text == parsed.Text

	if unchanged {
		fmt.Println("no changes")
		return nil
	}

	// Store.Update holds an exclusive lock across read-modify-write, so
	// concurrent appends that arrive between the editor session and this call
	// are included in the rewritten file rather than being silently discarded.
	var transformErr error
	updateErr := store.Update(func(entries []storage.Entry) []storage.Entry {
		// Validate inside the transform so that the check runs against the
		// locked snapshot, which may include entries appended since the
		// pre-editor read. resolvedRelated contains full IDs even when the
		// user typed prefixes in the edit buffer.
		resolvedRelated, err := validateEditedEntry(parsed, id, entries)
		if err != nil {
			transformErr = err
			// Returning the unmodified slice causes Update to rewrite the
			// file unchanged. We surface the error via the captured variable.
			return entries
		}

		for index, entry := range entries {
			if entry.ID == id {
				updated := entry
				updated.Topics = parsed.Topics
				updated.Related = resolvedRelated
				updated.Text = parsed.Text
				updated.UpdatedAt = time.Now().UTC()
				entries[index] = updated
				return entries
			}
		}
		// The entry was deleted between the editor session and the lock
		// acquisition. Surface this as a validation error.
		transformErr = fmt.Errorf("entry %q was deleted before the edit could be saved", id)
		return entries
	})
	if transformErr != nil {
		return transformErr
	}
	return updateErr
}
