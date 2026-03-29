### [Code Intelligence](/search_kw/8daff5985d71cb3ccef4c57533004215)

Prefer LSP over Grep/Read for code navigation — it's faster, precise, and avoids reading entire files:

- `workspaceSymbol` to find where something is defined
- `findReferences` to see all usages across the codebase
- `goToDefinition` / `goToImplementation` to jump to source
- `hover` for type info without reading the file

Use Grep only when LSP isn't available or for text/pattern searches (comments, strings, config).

After writing or editing code, check [LSP diagnostics](/search_kw/2c6f1042e8082ae61b79e6f7345ba173) and fix errors before proceeding.
