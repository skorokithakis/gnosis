// export_test.go exposes internal symbols for use by the external test package
// (package storage_test). Nothing in this file is compiled into production
// binaries.
package storage

import "time"

// StaleLockCutoff is the exported handle for staleLockCutoff so that tests can
// save and restore it to exercise both branches of removeStaleRepoLock.
var StaleLockCutoff = &staleLockCutoff

// RemoveStaleRepoLock calls the unexported removeStaleRepoLock method so that
// the external test package can invoke it directly without going through
// NewStore.
func (store *Store) RemoveStaleRepoLock() {
	store.removeStaleRepoLock()
}

// SetStaleLockCutoff replaces staleLockCutoff for the duration of a test and
// returns a restore function. Call it as: defer SetStaleLockCutoff(t, v)().
func SetStaleLockCutoff(cutoff time.Time) (restore func()) {
	original := staleLockCutoff
	staleLockCutoff = cutoff
	return func() { staleLockCutoff = original }
}
