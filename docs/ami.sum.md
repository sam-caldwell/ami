# Dependency Summary: `ami.sum`

JSON document mapping package → version → SHA‑256 commit identifier.

## Format

```
{
  "schema": "ami.sum/v1",
  "packages": {
    "github.com/example/foo": {
      "v1.2.3": "<sha256-commit-oid>"
    }
  }
}
```

## How It’s Produced

- `ami mod get git+ssh://…#vX.Y.Z`: resolves the tag and records the digest.
- `ami mod update`: resolves imports from `ami.workspace` and updates `ami.sum`.
- Writes are atomic (temp file then rename) with stable JSON ordering.

## Digests

- If the remote repository stores SHA‑256 commits, AMI records that OID.
- If the repository is SHA‑1, AMI computes a deterministic SHA‑256 over the raw commit object:
  - `"commit <len>\0" + <commit-bytes>` → SHA‑256 hex string.

## Verification

- `ami mod verify` re-computes digests for cached git checkouts and compares against `ami.sum`.
- `ami build` (future enforcement) will fail with exit code 3 on mismatch.

## `mod list --json`

- Includes the recorded digest for each cache entry when found in `ami.sum`.
