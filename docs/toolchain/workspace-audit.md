# Workspace Audit

Notes on workspace structure, dependency audit, and repository hygiene.

- Structure: `src/` for Go sources; `examples/` for AMI sample workspaces; `build/` for outputs.
- Dependency audit: prefer minimal, pinned versions.
- Determinism: stable JSON ordering, ISO-8601 UTC timestamps.
