package storage

import "os"

// fileLock wraps the open file descriptor the store uses as an advisory lock.
// It exists so the platform-specific acquire/release calls (flock on unix,
// LockFileEx on Windows) can live in build-tagged files while storage.go drives
// locking the same way on every platform.
type fileLock struct {
	file *os.File
}

// openFileLock opens (or creates) the lock file at the given path and returns a
// fileLock around it. It does not acquire any lock; the caller must call acquire
// to take the lock and must close the fileLock when done.
func openFileLock(path string) (*fileLock, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	return &fileLock{file: file}, nil
}

// close closes the underlying file descriptor. Callers must call the release
// function returned by acquire before calling close on platforms that need an
// explicit unlock (Windows); unix releases the flock implicitly on close, but
// the explicit unlock still runs there.
func (lock *fileLock) close() error {
	return lock.file.Close()
}
