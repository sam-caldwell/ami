# ami init

Initializes an AMI workspace in the current directory.

## Usage

- `ami init [--force]`

## Behavior

- Requires the current directory to be a Git repository. If it is not, the command fails with a user error unless `--force` is provided.
- With `--force`, a new Git repository is initialized (equivalent to `git init`).
- Writes a minimal `ami.workspace` scaffold at the repository root.
- Creates `src/main.ami` if it does not exist (or when `--force`).
- Ensures `.gitignore` contains `./build` (does not duplicate the entry; creates the file if missing).

### Scaffolded Files

- `ami.workspace`: minimal toolchain and package definition
- `src/main.ami`: a placeholder AMI source file with comments

See `docs/workspace.md` for the minimal `ami.workspace` schema and examples.

## Flags

- `--force`: Overwrite/initialize even when prerequisites are missing or files already exist.
- Global flags: `--json`, `--verbose`, `--color`, `--help` (see `docs/ami.md`).

## Output

- Human (default): informational messages printed to stdout; errors printed to stderr.
- JSON (`--json`): emits structured `diag.v1` records to stdout. Errors are still summarized to stderr for visibility.

## Exit Codes

- `0`: success
- `1`: user error (e.g., not a Git repo without `--force`, or refusing to overwrite without `--force`)
- See `docs/ami.md` for the standard exit code mapping.

## Examples

Initialize in an existing Git repo:

```
ami init
```

Initialize and create a Git repo if missing:

```
ami init --force
```

Emit machine‑readable output:

```
ami --json init --force
```

