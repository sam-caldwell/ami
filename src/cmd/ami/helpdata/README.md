AMI Help

Overview

- The `ami` CLI manages AMI workspaces, packages, linting, testing, and builds.
- Global flags: `--json`, `--verbose`, `--color` (mutually exclusive with `--json`).

Commands

- `ami help`             : show this help guide
- `ami version`          : print CLI version
- `ami init`             : create or update `ami.workspace`
- `ami clean`            : clean and recreate `./build`
- `ami lint`             : run linter checks
- `ami test`             : run project tests
- `ami build`            : build the workspace

Module commands (`ami mod *`)

- `ami mod clean`        : clean and recreate the package cache (`${AMI_PACKAGE_CACHE}`)
- `ami mod list`         : list cached packages (name, version, size, modified)
- `ami mod get <source>` : fetch a package into the cache and update `ami.sum`
   - Sources:
     - Local path: `./subproject` (must be within the workspace and declared in `ami.workspace`)
     - Git (non-interactive): `git+ssh://host/path#<tag>` or `file+git:///<abs-path>#<tag>`
- `ami mod update`       : copy workspace packages into the cache and refresh `ami.sum`
- `ami mod sum`          : validate `ami.sum` against the cache; fetches when `source` is set
- `ami mod audit`        : audit workspace requirements vs `ami.sum` and cache

Notes

- Human output goes to stdout; errors to stderr.
- JSON mode writes NDJSON for streaming commands.
- Use `--json` with module commands for machine-parsable output.
- See docs/Asynchronous Machine Interface.docx for the language and runtime semantics.
