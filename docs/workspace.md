# Workspace File: `ami.workspace`

Defines project metadata and dependencies for AMI.

## Minimal Scaffold

Created by `ami init`:

```
version: 1.0.0
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
```

## Dependencies

- Declare dependencies under `packages[].import`.
- Each import item is a repo path like `github.com/org/repo` and may be constrained (current phase supports either explicit `vX.Y.Z` or `==latest`).
- `ami mod update` uses these to fetch into the cache and write/update `ami.sum`.

Examples:

```
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - github.com/example/repo v1.2.3
        - github.com/other/lib ==latest
```

Notes:
- Local imports (e.g., `./subdir`) are allowed but not updated from remotes; they are copied into cache as `@local` when explicitly fetched.
- The file lives at the workspace root.
