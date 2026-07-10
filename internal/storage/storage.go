package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/skorokithakis/gnosis/internal/paths"
)

// IDAlphabet is the set of lowercase letters used for ID generation. The
// letters i, l, and o are excluded because they are visually confusable with 1,
// 1, and 0 respectively, which would make IDs harder to read and transcribe.
// It is exported so that callers (e.g. ID-prefix resolution in the commands
// package) can distinguish valid ID characters from other input without
// duplicating the alphabet.
const IDAlphabet = "abcdefghjkmnpqrstuvwxyz"

// IDLength is the number of characters in a generated entry ID. It is exported
// so that callers can use it to distinguish ID-prefix queries from topic queries
// without duplicating the constant.
const IDLength = 6

// Entry is the canonical in-memory representation of a knowledge entry. Topics
// always holds the normalized form (e.g. "keymaster-token-auth"), never the
// original casing the user typed. Normalization happens on write so that the
// stored data is always consistent.
type Entry struct {
	ID        string    `json:"id"`
	Topics    []string  `json:"topics"`
	Text      string    `json:"text"`
	Author    string    `json:"author,omitempty"`
	Related   []string  `json:"related"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store holds the resolved path to the .gnosis directory and provides all
// storage operations. Callers obtain a Store via NewStore.
type Store struct {
	gnosisDir string
	cacheDir  string
}

// NewStore resolves the repo root (by walking up from the current working
// directory) and returns a Store pointing at the .gnosis directory within it.
// The .gnosis directory is not created here; it is created on first write.
// The cache directory (for the lock file and index.db) is derived from the
// repo root via paths.CacheDir.
func NewStore() (*Store, error) {
	root, err := FindRepoRoot()
	if err != nil {
		return nil, err
	}
	cacheDir, err := paths.CacheDir(root)
	if err != nil {
		return nil, fmt.Errorf("resolving cache directory: %w", err)
	}
	return &Store{
		gnosisDir: filepath.Join(root, ".gnosis"),
		cacheDir:  cacheDir,
	}, nil
}

// NewStoreAt returns a Store that uses gnosisDir as its .gnosis directory and
// cacheDir as the directory for the lock file (and index.db). This is intended
// for tests that need to point the store at temporary directories rather than
// the real repo root.
func NewStoreAt(gnosisDir, cacheDir string) (*Store, error) {
	return &Store{gnosisDir: gnosisDir, cacheDir: cacheDir}, nil
}

// GnosisDir returns the path to the .gnosis directory this store uses.
func (store *Store) GnosisDir() string {
	return store.gnosisDir
}

// entriesPath returns the path to the JSONL entries file.
func (store *Store) entriesPath() string {
	return filepath.Join(store.gnosisDir, "entries.jsonl")
}

// lockPath returns the path to the lock file used to serialise rewrites.
// The lock file lives in the cache directory (alongside index.db), not in the
// repo, because it is purely runtime coordination state.
func (store *Store) lockPath() string {
	return filepath.Join(store.cacheDir, "lock")
}

// ensureDir creates the .gnosis directory if it does not already exist.
func (store *Store) ensureDir() error {
	return os.MkdirAll(store.gnosisDir, 0o755)
}

// ensureCacheDir creates the cache directory if it does not already exist.
func (store *Store) ensureCacheDir() error {
	return os.MkdirAll(store.cacheDir, 0o755)
}

// openLockFile opens (or creates) the lock file and returns a fileLock around
// it. The caller is responsible for acquiring the lock, releasing it, and
// closing the file.
func (store *Store) openLockFile() (*fileLock, error) {
	return openFileLock(store.lockPath())
}

// withSharedLock opens the lock file, acquires a shared lock, calls fn, then
// releases the lock and closes the file. A shared lock allows concurrent
// readers and appenders but blocks exclusive (rewrite) lockers. The cache
// directory is created if it does not exist, because the lock file lives there.
func (store *Store) withSharedLock(fn func() error) error {
	if err := store.ensureCacheDir(); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	lockFile, err := store.openLockFile()
	if err != nil {
		return fmt.Errorf("opening lock file: %w", err)
	}
	defer lockFile.close()

	release, err := lockFile.acquire(false)
	if err != nil {
		return fmt.Errorf("acquiring shared lock: %w", err)
	}
	defer release() //nolint:errcheck

	return fn()
}

// withExclusiveLock opens the lock file, acquires an exclusive lock, calls fn,
// then releases the lock and closes the file. An exclusive lock blocks all other
// lockers (both shared and exclusive) until it is released. The cache directory
// is created if it does not exist, because the lock file lives there.
func (store *Store) withExclusiveLock(fn func() error) error {
	if err := store.ensureCacheDir(); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	lockFile, err := store.openLockFile()
	if err != nil {
		return fmt.Errorf("opening lock file: %w", err)
	}
	defer lockFile.close()

	release, err := lockFile.acquire(true)
	if err != nil {
		return fmt.Errorf("acquiring exclusive lock: %w", err)
	}
	defer release() //nolint:errcheck

	return fn()
}

// Append normalizes all topics on entry, then serialises it as a single JSON
// line and appends it to the entries file. A shared lock on <cache-dir>/lock
// is held for the duration of the write so that a concurrent Rewrite (which
// takes an exclusive lock) cannot rename the file out from under us mid-write.
//
// Append takes a shared lock rather than an exclusive one because it assumes
// the caller has already assigned a unique ID, so it does not need to read the
// existing ID set. Production code should use AppendNew, which generates a
// collision-free ID under an exclusive lock; Append exists for callers that
// already hold a unique ID and is currently only used by tests.
//
// The shared lock coordinates Append against Rewrite but does not coordinate
// two concurrent Appends against each other. Their safety in practice rests on
// gn being a single-user CLI and on the OS appending whole short lines without
// interleaving, not on any regular-file write atomicity guarantee. (A previous
// version of this comment cited PIPE_BUF, but that atomicity only applies to
// pipes and FIFOs, so the claim did not actually hold for the entries file.)
func (store *Store) Append(entry Entry) error {
	if err := store.ensureDir(); err != nil {
		return fmt.Errorf("creating .gnosis directory: %w", err)
	}

	entry.Topics = normalizeTopics(entry.Topics)

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling entry: %w", err)
	}
	line = append(line, '\n')

	return store.withSharedLock(func() error {
		file, err := os.OpenFile(store.entriesPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("opening entries file: %w", err)
		}
		defer file.Close()

		if _, err := file.Write(line); err != nil {
			return fmt.Errorf("writing entry: %w", err)
		}
		return nil
	})
}

// AppendNew generates a collision-free ID under an exclusive lock, assigns it
// to entry, and appends the entry atomically. The exclusive lock serialises
// concurrent writers so that two goroutines cannot independently read the same
// ID set and generate the same ID. Plain Append (which takes a shared lock) is
// used when the caller already has a unique ID and only needs to block
// concurrent rewrites.
func (store *Store) AppendNew(entry Entry) (string, error) {
	if err := store.ensureDir(); err != nil {
		return "", fmt.Errorf("creating .gnosis directory: %w", err)
	}

	entry.Topics = normalizeTopics(entry.Topics)

	var assignedID string

	err := store.withExclusiveLock(func() error {
		existing, err := store.readAllUnlocked()
		if err != nil {
			return fmt.Errorf("reading existing entries: %w", err)
		}

		existingIDs := make(map[string]bool, len(existing))
		for _, existingEntry := range existing {
			existingIDs[existingEntry.ID] = true
		}

		assignedID = GenerateID(existingIDs)
		entry.ID = assignedID

		line, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("marshalling entry: %w", err)
		}
		line = append(line, '\n')

		file, err := os.OpenFile(store.entriesPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("opening entries file: %w", err)
		}
		defer file.Close()

		if _, err := file.Write(line); err != nil {
			return fmt.Errorf("writing entry: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return assignedID, nil
}

// ReadAll reads every entry from the JSONL file and returns them in order. If
// the file does not exist, an empty slice is returned without error, because an
// absent file is indistinguishable from an empty store.
func (store *Store) ReadAll() ([]Entry, error) {
	return store.readAllUnlocked()
}

// readAllUnlocked is the internal implementation of ReadAll. It does not
// acquire any lock, so callers that need locking must hold the appropriate lock
// before calling it.
func (store *Store) readAllUnlocked() ([]Entry, error) {
	file, err := os.Open(store.entriesPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("opening entries file: %w", err)
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)
	// The default 64 KB token limit would silently truncate or error on any
	// entry whose JSON line exceeds that size. 10 MB is generous enough for
	// human-written knowledge entries while still being a bounded allocation.
	const maxTokenSize = 10 * 1024 * 1024
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), maxTokenSize)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, fmt.Errorf("parsing entry at line %d: %w", lineNumber, err)
		}
		// Normalize topics on read so that legacy JSONL with non-normalized
		// topics self-heals without requiring an explicit migration step.
		entry.Topics = normalizeTopics(entry.Topics)
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading entries file: %w", err)
	}
	return entries, nil
}

// Rewrite atomically replaces the entire entries file with the given slice.
// It acquires an exclusive lock on <cache-dir>/lock before writing so that
// concurrent appenders cannot interleave with the rename. The write goes to a
// temp file in the same directory so that the final rename replaces the
// entries file in a single step rather than leaving a partially-written file.
func (store *Store) Rewrite(entries []Entry) error {
	if err := store.ensureDir(); err != nil {
		return fmt.Errorf("creating .gnosis directory: %w", err)
	}

	return store.withExclusiveLock(func() error {
		return store.rewriteUnlocked(entries)
	})
}

// rewriteUnlocked performs the actual temp-file write and atomic rename. The
// caller must hold the exclusive lock before calling this.
func (store *Store) rewriteUnlocked(entries []Entry) error {
	tmpFile, err := os.CreateTemp(store.gnosisDir, "entries-*.jsonl.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	writeErr := func() error {
		writer := bufio.NewWriter(tmpFile)
		for index, entry := range entries {
			entry.Topics = normalizeTopics(entry.Topics)
			line, err := json.Marshal(entry)
			if err != nil {
				return fmt.Errorf("marshalling entry %d: %w", index, err)
			}
			if _, err := writer.Write(line); err != nil {
				return fmt.Errorf("writing entry %d: %w", index, err)
			}
			if err := writer.WriteByte('\n'); err != nil {
				return fmt.Errorf("writing newline after entry %d: %w", index, err)
			}
		}
		return writer.Flush()
	}()

	if writeErr != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", writeErr)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, store.entriesPath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// Update takes an exclusive lock, reads all entries, passes them to transform,
// and atomically rewrites the result. The lock is held across the entire
// read-modify-write cycle so that no concurrent append or rewrite can
// interleave. This is the correct method for edit and rm operations.
func (store *Store) Update(transform func([]Entry) []Entry) error {
	if err := store.ensureDir(); err != nil {
		return fmt.Errorf("creating .gnosis directory: %w", err)
	}

	return store.withExclusiveLock(func() error {
		entries, err := store.readAllUnlocked()
		if err != nil {
			return fmt.Errorf("reading entries: %w", err)
		}

		updated := transform(entries)

		return store.rewriteUnlocked(updated)
	})
}

// WithSharedLock acquires a shared lock on <cache-dir>/lock and calls fn
// while holding it. This allows callers outside the storage package to perform
// multiple operations atomically under the same shared lock — for example,
// reading entries and then stat-ing the file mtime in a single critical section.
func (store *Store) WithSharedLock(fn func() error) error {
	return store.withSharedLock(fn)
}

// GenerateID returns a new random ID that is not present in existing. It retries
// until it finds a collision-free ID. The alphabet excludes i, l, and o to
// avoid visual confusion with digits.
func GenerateID(existing map[string]bool) string {
	for {
		id := make([]byte, IDLength)
		for position := range id {
			id[position] = IDAlphabet[rand.IntN(len(IDAlphabet))]
		}
		candidate := string(id)
		if !existing[candidate] {
			return candidate
		}
	}
}

// normalizeTopics returns a new slice with each topic normalized. It is a
// package-private helper used by Append, Rewrite, and ReadAll to ensure topics
// are always stored in canonical form.
func normalizeTopics(topics []string) []string {
	normalized := make([]string, len(topics))
	for index, topic := range topics {
		normalized[index] = NormalizeTopic(topic)
	}
	return normalized
}

// NormalizeTopic converts a display-form topic string into the canonical
// storage key used for lookups. The rules are:
//
//   - CamelCase boundaries (lower→upper or digit→upper) are split with a dash.
//   - Underscores and spaces are replaced with dashes.
//   - The result is lowercased.
//   - Consecutive dashes are collapsed to a single dash.
//   - Leading and trailing dashes are trimmed.
//
// This means "KeymasterTokenAuth", "keymaster_token_auth", and
// "keymaster-token-auth" all normalise to "keymaster-token-auth".
func NormalizeTopic(display string) string {
	runes := []rune(display)
	var builder strings.Builder

	for index, character := range runes {
		switch {
		case character == '_' || character == ' ':
			builder.WriteRune('-')
		case unicode.IsUpper(character) && index > 0:
			previous := runes[index-1]
			// Insert a dash before an uppercase letter at a CamelCase word
			// boundary. There are two boundary patterns:
			//
			// 1. lower/digit → upper: the standard CamelCase split, e.g.
			//    "TokenAuth" → "token-auth", "OAuth2Provider" → "…2-provider".
			//
			// 2. upper → upper followed by lower: handles acronyms where the
			//    last uppercase letter of the acronym starts the next word,
			//    e.g. "OAuth" → "o-auth" (the 'A' is upper, preceded by 'O'
			//    which is upper, and followed by 'u' which is lower).
			nextIsLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			if unicode.IsLower(previous) || unicode.IsDigit(previous) || (unicode.IsUpper(previous) && nextIsLower) {
				builder.WriteRune('-')
			}
			builder.WriteRune(unicode.ToLower(character))
		default:
			builder.WriteRune(unicode.ToLower(character))
		}
	}

	// Collapse consecutive dashes and strip leading/trailing ones.
	normalized := builder.String()
	parts := strings.FieldsFunc(normalized, func(r rune) bool { return r == '-' })
	return strings.Join(parts, "-")
}

// FindRepoRoot walks up from the current working directory looking for a .git
// directory, a .jj directory, or an existing .gnosis directory. The first
// ancestor directory that contains any of those markers is returned as the repo
// root. If no marker is found all the way to the filesystem root, the current
// working directory is returned so that gnosis can still be used outside of any
// version-controlled tree.
func FindRepoRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	directory := current
	for {
		for _, marker := range []string{".git", ".jj", ".gnosis"} {
			if _, err := os.Stat(filepath.Join(directory, marker)); err == nil {
				return directory, nil
			}
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			// Reached the filesystem root without finding a marker.
			return current, nil
		}
		directory = parent
	}
}
