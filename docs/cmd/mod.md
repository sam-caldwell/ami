# AMI Module Commands

This document describes the AMI module (mod) subcommands and their behavior.

Environment
- `AMI_PACKAGE_CACHE`: absolute path to the package cache. When unset, defaults to `${HOME}/.ami/pkg`. Commands create the directory if missing.
- Git operations set `GIT_TERMINAL_PROMPT=0` and use non‑interactive SSH. A short `-oConnectTimeout=2` is applied for deterministic failures.

Commands
- `ami mod clean`
  - Removes and recreates the package cache directory.
  - JSON: emits `{path, removed, created}`.
  - Human: prints `cleaned: <path>`.

- `ami mod list`
  - Lists cached packages and versions under the cache.
  - JSON: `{ path, entries: [{name, version, type, size, modified}] }`.
  - Human: `type\tname[@version]\tsize\t<ISO-8601-UTC>` per line.

- `ami mod get <src>`
  - Sources:
    - Local workspace path: `./<pkg-root>`; copies to cache using workspace `name@version`.
    - Git: `file+git://<abs-path>#<tag>` or `git+ssh://host/path#<tag>`.
      - If `#<tag>` omitted, selects the highest non‑prerelease SemVer tag.
  - Updates `ami.sum` in object form under the workspace root, recording `{version, sha256}` for the package.
  - JSON: `{source, name, version, path}`; network errors map to `NETWORK_REGISTRY_ERROR`.

- `ami mod sum`
  - Validates `ami.sum` schema and verifies cache contents match recorded hashes.
  - Attempts to fetch missing entries with an attached `source` when present; non‑fatal network issues result in integrity failure with `ok=false`.
  - JSON: `{schema, packages, verified, missing, mismatched, ok}`.

- `ami mod update`
  - Copies local workspace packages into the cache and refreshes `ami.sum` in canonical nested form.
  - JSON includes `audit` summary and `selected` highest satisfying versions for remote requirements (non‑destructive reporting).

Notes
- `ami clean` preserves `ami.sum` in the workspace root.
- All outputs are deterministic and non‑interactive; errors return stable exit codes.

