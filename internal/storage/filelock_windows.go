//go:build windows

package storage

import "golang.org/x/sys/windows"

// acquire takes a Windows byte-range lock on the open lock file. LockFileEx
// locks a range on an open handle rather than the file as a whole, so we lock
// the entire possible file (offset 0, length 2^64-1) to reproduce flock's
// whole-file semantics: any two lockers asking for the file overlap on this
// range and therefore conflict, regardless of which byte range each names.
//
// LOCKFILE_EXCLUSIVE_LOCK distinguishes exclusive from shared. A shared lock
// (the flag is clear) may overlap with other shared locks but conflicts with
// exclusive ones, matching flock's LOCK_SH; an exclusive lock conflicts with
// every other lock, matching LOCK_EX. LockFileEx blocks until the lock is
// granted unless LOCKFILE_FAIL_IMMEDIATELY is set, which we deliberately leave
// unset so the blocking behaviour matches flock's default and the call sites'
// assumption that success means the lock is held.
//
// Go opens files with FILE_SHARE_READ|FILE_SHARE_WRITE, so every process can
// keep the lock file open at once and LockFileEx is what actually serializes
// access (an O_WRONLY open would not, by itself, block a second opener on
// Windows). UnlockFileEx is invoked from the returned release function and must
// run before the handle is closed; relying on implicit unlock-on-close would
// leave the lock dangling if the deferred close were reordered relative to other
// work in the caller.
func (lock *fileLock) acquire(exclusive bool) (release func() error, err error) {
	var flags uint32
	if exclusive {
		flags = windows.LOCKFILE_EXCLUSIVE_LOCK
	}

	// Overlapped carries the lock's starting offset. Offset 0 means the locked
	// range begins at byte 0; the length is passed separately as low/high
	// halves below, so a zeroed Overlapped plus the maximum length covers the
	// whole file.
	var overlapped windows.Overlapped
	const lengthLow, lengthHigh uint32 = 0xffffffff, 0xffffffff

	handle := windows.Handle(lock.file.Fd())
	if err := windows.LockFileEx(handle, flags, 0, lengthLow, lengthHigh, &overlapped); err != nil {
		return nil, err
	}
	return func() error {
		return windows.UnlockFileEx(handle, 0, lengthLow, lengthHigh, &overlapped)
	}, nil
}
