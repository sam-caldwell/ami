# Registry Backends

AMI supports pluggable registry backends for fetching packages into the local cache (`${HOME}/.ami/pkg`).

Initial backends:

- `file`: local workspace paths declared in `ami.workspace`.
- `git+ssh`: Git repositories accessed over SSH with a required SemVer tag.

## How Resolution Works

- The `ami mod get <spec>` command selects a backend based on the spec:
  - `./path`, `../path`, `/abs/path`, or `file://…` → `file` backend.
  - `git+ssh://…#vX.Y.Z` → `git+ssh` backend.
- Each backend stages the module under `${HOME}/.ami/pkg/<name>@<version>` and logs the destination.
- When a backend returns a package name and a concrete semantic version (e.g., `git+ssh`), `ami mod get` 
  updates `ami.sum` with the commit digest.

## `file` Backend

- Accepts workspace‑relative or absolute paths, optionally prefixed with `file://`.
- Enforces safety:
  - Path must be inside the current workspace root (detected via `ami.workspace`).
  - Path must be declared in `packages.import` of `ami.workspace`.
- Copies the project directory into the cache as `<base>@local`.
- Does not update `ami.sum` directly (no tag). Use `ami mod update` with local repositories when you need versioned 
  digests.

Examples:
- `./build/ami mod get ./subproject`
- `./build/ami mod get file://./subproject`

## `git+ssh` Backend

- Accepts `git+ssh://host/org/repo.git#vX.Y.Z`.
- Uses your SSH agent for authentication (no interactive prompts).
- Clones shallowly, checks out the provided tag, and caches it as `<repo>@<tag>`.
- Updates `ami.sum` with a SHA‑256 of the raw commit object for the tag.

Examples:
- `./build/ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3`

## Extensibility

Backends implement a simple interface in `src/ami/mod` and are registered at init‑time. New schemes (e.g., `git+https`)
can be added without changing the CLI.

