package paths

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// CacheDir returns the directory that should hold per-repo runtime files
// (index.db, lock) for the given repo root. It respects XDG_CACHE_HOME; if
// that variable is unset it falls back to ~/.cache. The directory is keyed by
// the first 16 hex characters of sha256(repoRoot) so that multiple repos on
// the same machine each get their own cache directory without colliding.
func CacheDir(repoRoot string) (string, error) {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		base = filepath.Join(home, ".cache")
	}

	sum := sha256.Sum256([]byte(repoRoot))
	repoHash := fmt.Sprintf("%x", sum)[:16]

	return filepath.Join(base, "gnosis", repoHash), nil
}
