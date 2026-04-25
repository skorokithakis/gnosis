package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/skorokithakis/gnosis/internal/paths"
	"github.com/skorokithakis/gnosis/internal/storage"
)

// SearchHit is a single result returned by Search.
type SearchHit struct {
	EntryID string
	Snippet string
	Rank    float64
}

// Index wraps a SQLite FTS5 database that mirrors the JSONL entries file. It is
// disposable: if the database is deleted or corrupted, Rebuild recreates it
// from the JSONL source of truth.
type Index struct {
	db     *sql.DB
	store  *storage.Store
	dbPath string
}

// Open opens (or creates) the index database for the repo that owns store.
// repoRoot must be the absolute path returned by storage.FindRepoRoot.
func Open(repoRoot string, store *storage.Store) (*Index, error) {
	dir, err := paths.CacheDir(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("resolving cache directory: %w", err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	dbPath := filepath.Join(dir, "index.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening index database: %w", err)
	}

	// SQLite allows only one writer at a time; WAL mode improves concurrency
	// for the common read-heavy workload without requiring CGo.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if err := ensureSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensuring schema: %w", err)
	}

	return &Index{db: db, store: store, dbPath: dbPath}, nil
}

// ensureSchema creates the metadata table and the FTS5 virtual table if they
// do not already exist. The FTS5 table uses the porter stemmer so that
// "searching" matches "search", and unicode61 for correct Unicode tokenisation.
// prefix="2 3 4" enables prefix-index structures for 2-, 3-, and 4-character
// prefixes so that partial-word queries are fast.
func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS metadata (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);

		CREATE VIRTUAL TABLE IF NOT EXISTS entries_fts USING fts5(
			entry_id UNINDEXED,
			text,
			topics,
			tokenize = 'porter unicode61',
			prefix = '2 3 4'
		);
	`)
	return err
}

// EnsureFresh rebuilds the index if the JSONL file is newer than the mtime
// recorded in the metadata table, or if no mtime has been recorded yet.
func (index *Index) EnsureFresh() error {
	entriesPath := filepath.Join(index.store.GnosisDir(), "entries.jsonl")

	info, err := os.Stat(entriesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No JSONL file means nothing to index; the index is trivially fresh.
			return nil
		}
		return fmt.Errorf("stat entries file: %w", err)
	}
	currentMtime := info.ModTime().UTC().Format(time.RFC3339Nano)

	var storedMtime string
	row := index.db.QueryRow(`SELECT value FROM metadata WHERE key = 'mtime'`)
	if scanErr := row.Scan(&storedMtime); scanErr != nil && scanErr != sql.ErrNoRows {
		return fmt.Errorf("reading stored mtime: %w", scanErr)
	}

	if storedMtime == currentMtime {
		return nil
	}

	return index.Rebuild()
}

// Rebuild unconditionally drops and recreates the FTS5 table, loads all
// entries from the JSONL file, inserts them, and records the JSONL mtime.
//
// The shared flock is held across the ReadAll + mtime stat so that a concurrent
// append cannot land between the two operations and leave the index with a stale
// mtime that points to a snapshot missing the new entry. An exclusive (rewrite)
// lock blocks until the shared lock is released, so the mtime we record always
// matches the entries we read.
func (index *Index) Rebuild() error {
	entriesPath := filepath.Join(index.store.GnosisDir(), "entries.jsonl")

	// Read entries and capture the mtime under the same shared lock so that
	// the two observations are consistent. Without this, a concurrent append
	// could land between ReadAll and the stat, causing the index to record a
	// mtime that is newer than the entries it contains — EnsureFresh would
	// then consider the index fresh and skip the rebuild, leaving the new
	// entry unsearchable.
	var entries []storage.Entry
	var mtime string
	if err := index.store.WithSharedLock(func() error {
		var readErr error
		entries, readErr = index.store.ReadAll()
		if readErr != nil {
			return readErr
		}
		var mtimeErr error
		mtime, mtimeErr = jsonlMtime(entriesPath)
		return mtimeErr
	}); err != nil {
		return fmt.Errorf("reading entries under shared lock: %w", err)
	}

	tx, err := index.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(`DELETE FROM entries_fts`); err != nil {
		return fmt.Errorf("clearing FTS table: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO entries_fts (entry_id, text, topics) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing insert statement: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		topicsText := buildTopicsText(entry.Topics)
		if _, err := stmt.Exec(entry.ID, entry.Text, topicsText); err != nil {
			return fmt.Errorf("inserting entry %q: %w", entry.ID, err)
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO metadata (key, value) VALUES ('mtime', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		mtime,
	); err != nil {
		return fmt.Errorf("updating mtime metadata: %w", err)
	}

	return tx.Commit()
}

// buildTopicsText joins topics into a single space-separated string for FTS5
// indexing. Topics are already normalized at this point, so no further
// transformation is needed — the FTS5 tokenizer handles word splitting on
// dashes and lowercasing.
func buildTopicsText(topics []string) string {
	return strings.Join(topics, " ")
}

// jsonlMtime returns the modification time of the JSONL file formatted as
// RFC3339Nano. If the file does not exist it returns an empty string.
func jsonlMtime(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return info.ModTime().UTC().Format(time.RFC3339Nano), nil
}

// MatchStart and MatchEnd are sentinel bytes used as snippet() start/end
// markers. We use ASCII STX/ETX rather than printable characters so that the
// presentation layer (color vs. brackets) can be chosen by the caller without
// risk of collision with content that legitimately contains brackets.
const (
	MatchStart = "\x02"
	MatchEnd   = "\x03"
)

// Search runs an FTS5 query against the index and returns up to limit results
// ranked by BM25. The query string is passed through to FTS5 as-is, so callers
// may use FTS5 query syntax (phrase queries, column filters, prefix operators,
// etc.). Snippets are generated by FTS5's built-in snippet() function and
// delimit matched terms with MatchStart/MatchEnd sentinels.
func (index *Index) Search(query string, limit int) ([]SearchHit, error) {
	// snippet() arguments: table name, column index (-1 = best column),
	// start marker, end marker, ellipsis, number of tokens in snippet.
	rows, err := index.db.Query(`
		SELECT
			entry_id,
			snippet(entries_fts, -1, ?, ?, '…', 12),
			rank
		FROM entries_fts
		WHERE entries_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, MatchStart, MatchEnd, query, limit)
	if err != nil {
		return nil, fmt.Errorf("executing search query: %w", err)
	}
	defer rows.Close()

	var hits []SearchHit
	for rows.Next() {
		var hit SearchHit
		if err := rows.Scan(&hit.EntryID, &hit.Snippet, &hit.Rank); err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}
		hits = append(hits, hit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search results: %w", err)
	}
	return hits, nil
}

// DBPath returns the absolute path to the SQLite database file backing this
// index. Callers can use this to report the cache location to the user without
// needing to know how the path is derived.
func (index *Index) DBPath() string {
	return index.dbPath
}

// Close closes the underlying database connection.
func (index *Index) Close() error {
	return index.db.Close()
}
