# GitHub Actions CI Design

**Date:** 2026-03-29

## Overview

We added a GitHub Actions CI workflow to the mud project that runs all quality checks defined in `.pre-commit-config.yaml`. Rather than duplicating checks as native shell commands, the CI uses the `prek-action` which runs all pre-commit hooks directly.

## Design Decisions

### Single Source of Truth

All quality checks are defined in `.pre-commit-config.yaml`. This ensures that:

- Local development (`pre-commit` hooks) and CI (GitHub Actions) run the same checks
- Developers can test changes locally before pushing
- No drift between local and CI quality standards

### Simplified CI Configuration

The `.github/workflows/ci.yml` workflow is minimal and delegates all work to:

- `actions/setup-go@v5` - sets up Go (required by go-fmt, go-vet, go-unit-tests)
- `j178/prek-action@v2` - runs all pre-commit hooks

Node.js is pre-installed on `ubuntu-latest`, so `npx prettier@3.8.1` works without additional setup.

## Checks Performed

All hooks from `.pre-commit-config.yaml`:

1. **File hygiene** - `check-toml`, `check-yaml`, `end-of-file-fixer`, `mixed-line-ending`, `trailing-whitespace`
2. **Go checks** - `go-fmt`, `go-vet`, `go-unit-tests`
3. **Code formatting** - `prettier@3.8.1 --write` (fails if any file is modified)
4. **Secret scanning** - `gitleaks`

## Implementation Details

### `.pre-commit-config.yaml` Changes

Updated the local prettier hook to use a pinned version via `npx`:

```yaml
- repo: local
  hooks:
    - id: prettier
      name: prettier
      entry: npx prettier@3.8.1 --write
      language: system
      types_or: [json, markdown, yaml]
```

Benefits:

- No global prettier installation required
- Pinned version (3.8.1) ensures consistent formatting across all environments
- Works on any machine with Node.js

### `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
  pull_request:

jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - uses: j178/prek-action@v2
```

The workflow:

- Checks out the code
- Sets up Go 1.26
- Runs all pre-commit hooks via `prek-action`

## Verification

To verify the CI works:

1. Push a branch to GitHub
2. Check the Actions tab to confirm the `CI` workflow appears
3. Verify all prek steps pass
4. Optionally introduce a formatting violation to confirm CI catches it
