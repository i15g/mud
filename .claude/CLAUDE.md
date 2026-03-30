- @README.md
- @docs/development.md

### Claude Hooks

- **post-tool-use-go-test** — after editing a `.go` source file, runs `go test` in that package and surfaces failures as context
- **post-tool-use-prek** — after any file edit, runs `prek run --files <path>` to auto-apply pre-commit hooks

## Code Intelligence

Prefer LSP over Grep/Read for code navigation — it's faster, precise, and avoids reading entire files:

- `workspaceSymbol` to find where something is defined
- `findReferences` to see all usages across the codebase
- `goToDefinition` / `goToImplementation` to jump to source
- `hover` for type info without reading the file

Use Grep only when LSP isn't available or for text/pattern searches (comments, strings, config).

After writing or editing code, check LSP diagnostics and fix errors before proceeding.
