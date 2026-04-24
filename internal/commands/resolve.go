package commands

import (
	"fmt"
	"strings"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// idAlphabetSet is the set of characters that may appear in an entry ID. It
// mirrors the alphabet constant in the storage package. Characters outside this
// set cannot be part of any ID, so a prefix containing them can never match.
const idAlphabetSet = "abcdefghjkmnpqrstuvwxyz"

// isIDAlphabetChar reports whether character is in the ID alphabet.
func isIDAlphabetChar(character rune) bool {
	return strings.ContainsRune(idAlphabetSet, character)
}

// ResolveIDPrefix finds the single entry whose ID starts with prefix and
// returns its full ID. Errors are returned for two distinct failure modes:
//
//   - Not found: no entry ID starts with prefix, or prefix contains characters
//     outside the ID alphabet (which guarantees no match is possible).
//   - Ambiguous: more than one entry ID starts with prefix; the error message
//     lists all candidate full IDs so the caller can report them to the user.
//
// A full 6-character ID is the degenerate case of a unique prefix.
func ResolveIDPrefix(entries []storage.Entry, prefix string) (string, error) {
	// An empty prefix would match every entry, which is never a useful
	// operation and would produce a misleading "ambiguous" error in a
	// multi-entry store or silently resolve in a single-entry store.
	if prefix == "" {
		return "", fmt.Errorf("entry %q not found", prefix)
	}

	// A prefix containing characters outside the ID alphabet can never match
	// any entry ID, so we treat it as not found rather than scanning all entries.
	for _, character := range prefix {
		if !isIDAlphabetChar(character) {
			return "", fmt.Errorf("entry %q not found", prefix)
		}
	}

	var candidates []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.ID, prefix) {
			candidates = append(candidates, entry.ID)
		}
	}

	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("entry %q not found", prefix)
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("prefix %q is ambiguous: matches %s", prefix, strings.Join(candidates, ", "))
	}
}
