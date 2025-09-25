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
- Each import item is a repo path like `github.com/org/repo` and may include a version constraint.
- Supported constraint forms:
  - Exact: `1.2.3` or `v1.2.3` (SemVer `MAJOR.MINOR.PATCH`)
  - Caret: `^1.2.3` (compatible with the same major)
  - Tilde: `~1.2.3` (compatible within the same minor)
  - Greater-than: `>1.2.3`
  - Greater-than-or-equal: `>=1.2.3`
  - Macro latest: `==latest`
- Notes:
  - Whitespace inside constraints is ignored; `>= 1.2.3` is accepted.
  - Prereleases (e.g., `-rc.1`) are allowed when specified in the constraint; otherwise they are excluded by default.
  - Unsupported operators (e.g., `<=`) are rejected.
- `ami mod update` uses these constraints to fetch into the cache and write/update `ami.sum`.

Examples:

```
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - github.com/example/repo v1.2.3
        - github.com/compat/lib ^1.2.0
        - github.com/patch/lib ~1.2.3
        - github.com/range/lib >=1.0.0
        - github.com/strict/lib >1.2.3
        - github.com/rolling/lib ==latest
        - ./local/module ==latest
```

Notes:
- Local imports (e.g., `./subdir`) are allowed and treated as workspace-local repositories; when used with `==latest`, the latest non-prerelease tag is selected from the local repo. They are copied into the cache under `<name>@<tag>`.
- The file lives at the workspace root.
