---
title: GitHub Actions CI Plan
status: ready
---

## Context

The project has no CI. We want to add a basic GitHub Actions workflow. The project already uses `prek` with a `.pre-commit-config.yaml` defining all quality checks (go-fmt, go-vet, go-unit-tests, prettier, gitleaks, file hygiene). Rather than duplicating these checks as native shell commands, CI should run `prek` directly so the pre-commit config remains the single source of truth.

## Files to Create/Modify

- **Create:** `.github/workflows/ci.yml`
- **Modify:** `.pre-commit-config.yaml` (pin prettier version, switch to `npx`)
- **Create:** `docs/superpowers/specs/2026-03-29-github-actions-ci-design.md`

## Changes

### `.pre-commit-config.yaml`

Change the local prettier hook to use `npx prettier@3.8.1` (pinned version, no global install required):

```yaml
- repo: local
  hooks:
    - id: prettier
      name: prettier
      entry: npx prettier@3.8.1 --write
      language: system
      types_or: [json, markdown, yaml]
```

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

`setup-go` is required before prek because the Go hooks (go-fmt, go-vet, go-unit-tests) need Go available. Node.js is pre-installed on ubuntu-latest so `npx` works without extra setup.

## What prek runs in CI

All hooks from `.pre-commit-config.yaml`:

- `check-toml`, `check-yaml`, `end-of-file-fixer`, `mixed-line-ending`, `trailing-whitespace`
- `go-fmt`, `go-vet`, `go-unit-tests`
- `prettier@3.8.1 --write` (fails if any file is modified)
- `gitleaks` secret scan

## Verification

Push the branch to GitHub and confirm:

1. The `CI` workflow appears under Actions
2. All prek steps pass
3. A deliberate formatting change triggers a failure
