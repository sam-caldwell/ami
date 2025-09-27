AMI Help

Overview

- The `ami` CLI manages AMI workspaces, packages, linting, testing, and builds.
- Global flags: `--json`, `--verbose`, `--color` (mutually exclusive with `--json`).

Commands

- `ami help`         : show this help guide
- `ami version`      : print CLI version
- `ami init`         : create or update `ami.workspace`
- `ami clean`        : clean and recreate `./build`
- `ami mod clean`    : clean and recreate the package cache
- `ami mod list`     : list cached packages
- `ami mod audit`    : audit workspace imports vs `ami.sum` and cache
- `ami mod sum`      : validate cache against `ami.sum`
- `ami lint`         : run linter checks
- `ami test`         : run project tests
- `ami build`        : build the workspace

Notes

- Human output goes to stdout; errors to stderr.
- JSON mode writes NDJSON for streaming commands.
- `ami mod update` surfaces a non-fatal audit summary before updates. In human mode it prints lines prefixed with `audit:`; in `--json` mode it includes an `audit` object alongside `updated`.
- See docs/Asynchronous Machine Interface.docx for the language and runtime semantics.
