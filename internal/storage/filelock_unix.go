//go:build unix

package storage

import "golang.org/x/sys/unix"

// acquire takes a flock on the open lock file. When exclusive is false the lock
// is shared (LOCK_SH), allowing concurrent readers and appenders but blocking
// exclusive rewriters; when true it is exclusive (LOCK_EX), blocking all other
// lockers until it is released. Like flock itself, acquire blocks until the lock
// can be granted, which matches the call sites' assumption that success means
// the lock is held.
//
// We use golang.org/x/sys/unix rather than the standard library syscall package
// so that this file can carry the broad //go:build unix tag: stdlib syscall has
// no Flock on solaris, whereas x/sys/unix exposes it for every unix port.
//
// The returned release function drops the lock with LOCK_UN. The kernel would
// also drop it when the descriptor closes, but releasing explicitly keeps the
// release path identical across platforms and avoids holding the lock for the
// remainder of any deferred close.
func (lock *fileLock) acquire(exclusive bool) (release func() error, err error) {
	how := unix.LOCK_SH
	if exclusive {
		how = unix.LOCK_EX
	}
	if err := unix.Flock(int(lock.file.Fd()), how); err != nil {
		return nil, err
	}
	return func() error {
		return unix.Flock(int(lock.file.Fd()), unix.LOCK_UN)
	}, nil
}
