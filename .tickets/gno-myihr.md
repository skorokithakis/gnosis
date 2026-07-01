---
id: gno-myihr
status: closed
deps: [gno-cyvqa]
links: []
created: 2026-07-01T21:22:45Z
type: task
priority: 2
assignee: Stavros Korokithakis
---
# Windows support: portable file locking and release targets

Ready for implementation.

Objective: make gn build and work on Windows.

Scope:
- internal/storage/storage.go currently calls syscall.Flock directly (unix-only). Extract the lock/unlock operations into build-tagged files (filelock_unix.go with a unix build tag, filelock_windows.go using golang.org/x/sys/windows LockFileEx/UnlockFileEx, with LOCKFILE_EXCLUSIVE_LOCK distinguishing exclusive from shared). Keep the withSharedLock/withExclusiveLock call sites unchanged.
- Prefer golang.org/x/sys (already an indirect dependency) over adding a new third-party locking library.
- Add windows (amd64, arm64) to the goos matrix in .goreleaser.yaml. Use zip archive format for windows if goreleaser conventions call for it.
- Verify cross-compilation: GOOS=windows go build ./... must succeed, and go test ./... must still pass on linux.

Non-goals: no Windows CI runner, no Windows-specific testing beyond cross-compilation, no changes to cache-dir resolution (os.UserHomeDir handles Windows), no installer/scoop/chocolatey packaging.

## Design

Hand-rolled build-tag shim over a locking dependency: the repo culture is dependency-averse (Bleve was rejected partly on transitive-dep count), the shim is small, and x/sys is already in the module graph. Windows support is best-effort per the owner; do not contort the design for it.

