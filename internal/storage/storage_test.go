package storage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skorokithakis/gnosis/internal/storage"
)

// --- Lock file location ---

// TestLockFile_lands_in_cacheDir verifies that after a write operation the lock
// file exists under cacheDir (not under gnosisDir).
func TestLockFile_lands_in_cacheDir(t *testing.T) {
	gnosisDir := filepath.Join(t.TempDir(), ".gnosis")
	cacheDir := t.TempDir()
	store, err := storage.NewStoreAt(gnosisDir, cacheDir)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}

	if err := store.Append(sampleEntry("aaaaaa")); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// The lock file must exist in cacheDir.
	lockPath := filepath.Join(cacheDir, "lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Errorf("expected lock file at %q, got error: %v", lockPath, err)
	}

	// The old location inside gnosisDir must not exist.
	oldLockPath := filepath.Join(gnosisDir, ".lock")
	if _, err := os.Stat(oldLockPath); err == nil {
		t.Errorf("lock file must not exist at old location %q", oldLockPath)
	}
}

// --- Stale-lock cleanup ---

// TestRemoveStaleRepoLock_removes_before_cutoff verifies that
// RemoveStaleRepoLock deletes <gnosisDir>/.lock when called before the cutoff.
func TestRemoveStaleRepoLock_removes_before_cutoff(t *testing.T) {
	defer storage.SetStaleLockCutoff(time.Now().Add(24 * time.Hour))()

	gnosisDir := filepath.Join(t.TempDir(), ".gnosis")
	if err := os.MkdirAll(gnosisDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	staleLock := filepath.Join(gnosisDir, ".lock")
	if err := os.WriteFile(staleLock, nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store, err := storage.NewStoreAt(gnosisDir, t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	store.RemoveStaleRepoLock()

	if _, err := os.Stat(staleLock); err == nil {
		t.Errorf("stale lock file should have been removed but still exists at %q", staleLock)
	}
}

// TestRemoveStaleRepoLock_skips_after_cutoff verifies that RemoveStaleRepoLock
// leaves <gnosisDir>/.lock untouched when called after the cutoff.
func TestRemoveStaleRepoLock_skips_after_cutoff(t *testing.T) {
	defer storage.SetStaleLockCutoff(time.Now().Add(-24 * time.Hour))()

	gnosisDir := filepath.Join(t.TempDir(), ".gnosis")
	if err := os.MkdirAll(gnosisDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	staleLock := filepath.Join(gnosisDir, ".lock")
	if err := os.WriteFile(staleLock, nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store, err := storage.NewStoreAt(gnosisDir, t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	store.RemoveStaleRepoLock()

	if _, err := os.Stat(staleLock); err != nil {
		t.Errorf("stale lock file should not have been removed after cutoff, but got: %v", err)
	}
}

// --- gnosisDir not created by read-only operations ---

// TestReadAll_does_not_create_gnosisDir verifies that ReadAll on a fresh store
// (no entries file) does not create the .gnosis directory. The directory should
// only be created when something is actually written to entries.jsonl.
func TestReadAll_does_not_create_gnosisDir(t *testing.T) {
	gnosisDir := filepath.Join(t.TempDir(), ".gnosis")
	store, err := storage.NewStoreAt(gnosisDir, t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}

	if _, err := os.Stat(gnosisDir); err == nil {
		t.Errorf("ReadAll must not create gnosisDir %q", gnosisDir)
	}
}

// newTestStore creates a Store backed by a temporary directory and returns both
// the store and a cleanup function. Using a temp directory isolates each test
// from the real filesystem and from other tests.
func newTestStore(t *testing.T) *storage.Store {
	t.Helper()
	tempDir := t.TempDir()
	store, err := storage.NewStoreAt(filepath.Join(tempDir, ".gnosis"), t.TempDir())
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
		Topics:    []string{"GoLang", "testing"},
		Text:      "some text",
		Related:   []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// --- ID generation ---

func TestGenerateID_length(t *testing.T) {
	id := storage.GenerateID(map[string]bool{})
	if len(id) != 6 {
		t.Errorf("expected ID length 6, got %d (%q)", len(id), id)
	}
}

func TestGenerateID_no_confusable_letters(t *testing.T) {
	confusable := map[rune]bool{'i': true, 'l': true, 'o': true}
	for range 1000 {
		id := storage.GenerateID(map[string]bool{})
		for _, character := range id {
			if confusable[character] {
				t.Errorf("ID %q contains confusable letter %q", id, character)
			}
		}
	}
}

func TestGenerateID_only_lowercase_letters(t *testing.T) {
	for range 1000 {
		id := storage.GenerateID(map[string]bool{})
		for _, character := range id {
			if character < 'a' || character > 'z' {
				t.Errorf("ID %q contains non-lowercase-letter character %q", id, character)
			}
		}
	}
}

func TestGenerateID_retries_on_collision(t *testing.T) {
	// Fill the existing map with all possible IDs except one, then verify that
	// GenerateID eventually returns the only remaining valid ID. Because the
	// alphabet has 23 characters and IDs are 6 characters long, the total
	// space is 23^6 = 148,035,889 IDs, which is far too large to enumerate.
	// Instead we pre-populate a set with all IDs except a single known one and
	// confirm that GenerateID returns that one.
	//
	// We use a tiny synthetic alphabet for this test by constructing the
	// existing map to contain every 6-char string from the real alphabet
	// except "aaaaaa". This is still too large, so we instead test the retry
	// logic by providing a map that contains the first several IDs that the
	// PRNG would generate. Since we cannot predict the PRNG output, we instead
	// verify the simpler invariant: GenerateID never returns an ID that is in
	// the existing map.
	existing := map[string]bool{
		"aaaaaa": true,
		"bbbbbb": true,
		"cccccc": true,
	}
	for range 100 {
		id := storage.GenerateID(existing)
		if existing[id] {
			t.Errorf("GenerateID returned a colliding ID %q", id)
		}
	}
}

// --- Topic normalisation ---

func TestNormalizeTopic_camel_case(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"KeymasterTokenAuth", "keymaster-token-auth"},
		{"keymaster_token_auth", "keymaster-token-auth"},
		{"keymaster-token-auth", "keymaster-token-auth"},
		{"keymaster token auth", "keymaster-token-auth"},
		// Mixed input: CamelCase with underscores.
		{"Keymaster_TokenAuth", "keymaster-token-auth"},
		// All lowercase, no separators.
		{"golang", "golang"},
		// Single word, capitalised.
		{"Golang", "golang"},
		// Acronym followed by word: the transition from the last uppercase of
		// the acronym to the next uppercase starts a new segment only when
		// followed by a lowercase letter. "OAuth2" should stay as one segment
		// because "2" is a digit, not a letter boundary.
		{"OAuth2", "o-auth2"},
		// Digit followed by uppercase triggers a split.
		{"OAuth2Provider", "o-auth2-provider"},
		// Already normalised.
		{"already-normalised", "already-normalised"},
		// Repeated separators collapse.
		{"foo--bar", "foo-bar"},
		{"foo__bar", "foo-bar"},
		{"foo  bar", "foo-bar"},
	}

	for _, testCase := range cases {
		result := storage.NormalizeTopic(testCase.input)
		if result != testCase.expected {
			t.Errorf("NormalizeTopic(%q) = %q, want %q", testCase.input, result, testCase.expected)
		}
	}
}

// --- Topic normalization on write ---

// TestAppend_normalizes_topics verifies that Append stores topics in normalized
// form regardless of how they were originally written.
func TestAppend_normalizes_topics(t *testing.T) {
	store := newTestStore(t)

	entry := sampleEntry("aaaaaa")
	entry.Topics = []string{"KeymasterTokenAuth", "keymaster_token_auth", "BILLING"}
	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	got, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	// "KeymasterTokenAuth" and "keymaster_token_auth" both normalize to
	// "keymaster-token-auth", but Append does not deduplicate — that is the
	// caller's responsibility. Both are stored in normalized form.
	if len(got[0].Topics) != 3 {
		t.Fatalf("expected 3 topics, got %v", got[0].Topics)
	}
	if got[0].Topics[0] != "keymaster-token-auth" {
		t.Errorf("topic 0: expected %q, got %q", "keymaster-token-auth", got[0].Topics[0])
	}
	if got[0].Topics[1] != "keymaster-token-auth" {
		t.Errorf("topic 1: expected %q, got %q", "keymaster-token-auth", got[0].Topics[1])
	}
	if got[0].Topics[2] != "billing" {
		t.Errorf("topic 2: expected %q, got %q", "billing", got[0].Topics[2])
	}
}

// TestReadAll_normalizes_legacy_topics verifies that ReadAll normalizes topics
// from entries that were written before normalization was enforced, so that
// legacy JSONL files self-heal on the next read without a migration step.
func TestReadAll_normalizes_legacy_topics(t *testing.T) {
	store := newTestStore(t)

	// Write a raw JSONL line with non-normalized topics, bypassing Append so
	// that the normalization in Append does not interfere with the test.
	gnosisDir := store.GnosisDir()
	if err := os.MkdirAll(gnosisDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	rawLine := `{"id":"aaaaaa","topics":["KeymasterTokenAuth","BILLING"],"text":"t","related":[],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(gnosisDir, "entries.jsonl"), []byte(rawLine), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Topics[0] != "keymaster-token-auth" {
		t.Errorf("topic 0: expected %q, got %q", "keymaster-token-auth", got[0].Topics[0])
	}
	if got[0].Topics[1] != "billing" {
		t.Errorf("topic 1: expected %q, got %q", "billing", got[0].Topics[1])
	}
}

// --- JSONL round-trip ---

func TestRoundTrip(t *testing.T) {
	store := newTestStore(t)

	entries := []storage.Entry{
		sampleEntry("aaaaaa"),
		sampleEntry("bbbbbb"),
		sampleEntry("cccccc"),
	}
	entries[1].Topics = []string{"Different", "topics"}
	entries[1].Text = "different text"
	entries[2].Related = []string{"aaaaaa"}

	for _, entry := range entries {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}

	for index, entry := range entries {
		if got[index].ID != entry.ID {
			t.Errorf("entry %d: ID mismatch: got %q, want %q", index, got[index].ID, entry.ID)
		}
		if got[index].Text != entry.Text {
			t.Errorf("entry %d: Text mismatch: got %q, want %q", index, got[index].Text, entry.Text)
		}
		if !got[index].CreatedAt.Equal(entry.CreatedAt) {
			t.Errorf("entry %d: CreatedAt mismatch: got %v, want %v", index, got[index].CreatedAt, entry.CreatedAt)
		}
		if !got[index].UpdatedAt.Equal(entry.UpdatedAt) {
			t.Errorf("entry %d: UpdatedAt mismatch: got %v, want %v", index, got[index].UpdatedAt, entry.UpdatedAt)
		}
	}
}

func TestReadAll_empty_store(t *testing.T) {
	store := newTestStore(t)
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on empty store: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestRewrite_replaces_all(t *testing.T) {
	store := newTestStore(t)

	original := []storage.Entry{sampleEntry("aaaaaa"), sampleEntry("bbbbbb")}
	for _, entry := range original {
		if err := store.Append(entry); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	replacement := []storage.Entry{sampleEntry("cccccc")}
	if err := store.Rewrite(replacement); err != nil {
		t.Fatalf("Rewrite: %v", err)
	}

	got, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after Rewrite: %v", err)
	}
	if len(got) != 1 || got[0].ID != "cccccc" {
		t.Errorf("expected single entry cccccc after Rewrite, got %v", got)
	}
}

// --- Concurrent appends ---

func TestConcurrentAppend(t *testing.T) {
	store := newTestStore(t)

	const goroutineCount = 20
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for index := range goroutineCount {
		go func(index int) {
			defer waitGroup.Done()
			entry := sampleEntry(fmt.Sprintf("a%05d", index)[:6])
			if err := store.Append(entry); err != nil {
				t.Errorf("goroutine %d: Append: %v", index, err)
			}
		}(index)
	}

	waitGroup.Wait()

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after concurrent appends: %v", err)
	}
	if len(entries) != goroutineCount {
		t.Errorf("expected %d entries after concurrent appends, got %d", goroutineCount, len(entries))
	}
}

// --- Concurrent append + update (regression for data-loss bug) ---

// TestConcurrentAppendAndUpdate is the regression test for the data-loss bug
// where concurrent appends were silently discarded by a concurrent rewrite.
// Before the fix, Rewrite held the exclusive lock only during the rename, so
// appends that arrived between ReadAll and the rename went to the old file and
// were overwritten. After the fix, Update holds the exclusive lock across the
// entire read-modify-write cycle, and Append holds a shared lock for the
// duration of the write, so the two operations are properly serialised.
//
// This test MUST fail against the old code (where Append has no lock and
// Rewrite only locks during the rename) and pass against the new code.
func TestConcurrentAppendAndUpdate(t *testing.T) {
	store := newTestStore(t)

	const appenderCount = 5
	const rewriteCount = 3

	// Seed the store with one entry so Update has something to work with.
	seed := sampleEntry("seed00")
	if err := store.Append(seed); err != nil {
		t.Fatalf("seeding store: %v", err)
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(appenderCount + rewriteCount)

	// appendedIDs records the IDs that each appender successfully wrote.
	appendedIDs := make([]string, appenderCount)

	for index := range appenderCount {
		go func(index int) {
			defer waitGroup.Done()
			id := fmt.Sprintf("a%05d", index)[:6]
			appendedIDs[index] = id
			entry := sampleEntry(id)
			if err := store.Append(entry); err != nil {
				t.Errorf("appender %d: Append: %v", index, err)
			}
		}(index)
	}

	for index := range rewriteCount {
		go func(index int) {
			defer waitGroup.Done()
			// Update performs a no-op transform (returns entries unchanged).
			// Its purpose here is to exercise the exclusive-lock path and
			// verify that it does not discard concurrent appends.
			if err := store.Update(func(entries []storage.Entry) []storage.Entry {
				return entries
			}); err != nil {
				t.Errorf("updater %d: Update: %v", index, err)
			}
		}(index)
	}

	waitGroup.Wait()

	final, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after concurrent test: %v", err)
	}

	finalIDs := make(map[string]bool, len(final))
	for _, entry := range final {
		finalIDs[entry.ID] = true
	}

	// Every appended ID must survive in the final state.
	for index, id := range appendedIDs {
		if !finalIDs[id] {
			t.Errorf("appender %d: entry %q was lost (data-loss bug)", index, id)
		}
	}
}

// TestRoundTrip_large_body verifies that entries whose JSON line exceeds the
// default 64 KB bufio.Scanner token limit survive a full write-read cycle.
// Before the scanner.Buffer fix, any entry with a body over ~64 KB would cause
// ReadAll to return a "token too long" error.
func TestRoundTrip_large_body(t *testing.T) {
	store := newTestStore(t)

	// 2 MB of repeated text — well above the 64 KB default scanner limit.
	const bodySize = 2 * 1024 * 1024
	largeText := strings.Repeat("x", bodySize)

	entry := sampleEntry("aaaaaa")
	entry.Text = largeText

	if err := store.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	got, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Text != largeText {
		t.Errorf("Text mismatch: got %d bytes, want %d bytes", len(got[0].Text), len(largeText))
	}
}

// --- FindRepoRoot ---

func TestFindRepoRoot_finds_git(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatalf("creating .git: %v", err)
	}

	subDir := filepath.Join(tempDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("creating subdir: %v", err)
	}

	// Change working directory to the subdirectory so FindRepoRoot walks up.
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(original) //nolint:errcheck

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	root, err := storage.FindRepoRoot()
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}
	if root != tempDir {
		t.Errorf("expected root %q, got %q", tempDir, root)
	}
}

func TestFindRepoRoot_falls_back_to_cwd(t *testing.T) {
	// Use a temp directory with no markers so FindRepoRoot falls back to CWD.
	tempDir := t.TempDir()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	defer os.Chdir(original) //nolint:errcheck

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	root, err := storage.FindRepoRoot()
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}
	if root != tempDir {
		t.Errorf("expected root %q, got %q", tempDir, root)
	}
}
