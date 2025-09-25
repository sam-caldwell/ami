# ami mod

Module and cache operations.

## clean

- Recreate the package cache directory `${HOME}/.ami/pkg`.
- Intended for removing stale/corrupt caches.

Usage:
- `ami mod clean`

## update

- Resolve and fetch all dependencies declared in `ami.workspace` under `packages[].import`.
- Supports constraints on each import item:
  - Exact: `1.2.3` or `v1.2.3`
  - Caret: `^1.2.3`
  - Tilde: `~1.2.3`
  - Greater-than: `>1.2.3`
  - Greater-than-or-equal: `>=1.2.3`
  - Macro latest: `==latest`
- Updates `ami.sum` with the resolved version → digest mapping.

Usage:
- `ami mod update`

## get

- Fetch a single package into the cache.
- Supported URL forms (initial):
  - `git+ssh://git@host/org/repo.git#vX.Y.Z` (required semver tag)
  - Local path: `./subproject` (copied into cache; version is `local`)
- On success, logs the destination path and updates `ami.sum` when a tag is provided.

Usage:
- `ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3`

## list

- Human: one line per entry in the form `<name>@<version>`.
- JSON (`--json`): one `diag.v1` record per entry with message `"cache.entry"` and `data`:
  - `entry`: `<name>@<version>`
  - `name`: base repository name
  - `version`: semantic version or `local`
  - `digest`: recorded digest from `ami.sum` (present only if found)

Notes:
- Reads `ami.sum` from the current directory to include digests when available.
- Output order is stable (entries sorted lexicographically).

Usage:
- `ami mod list`
- `ami mod list --json`

## verify

- Verifies `ami.sum` against the local cache:
  - Ensures each cached entry exists.
  - Recomputes the commit digest for git checkouts and compares against `ami.sum`.
- Logs errors for mismatches or missing entries.

Usage:
- `ami mod verify`
