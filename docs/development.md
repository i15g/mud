# Development

## mise

- https://mise.jdx.dev/ (tool version manager)
- `mise.toml` pins go, gopls, node, prettier, prek; `mise.lock` (+ `.mise/locks/`) holds checksums. Commit all
- prek hooks call tools via `mise exec --`, so they use these pins regardless of PATH

```sh
mise trust               # once per clone
mise install --locked    # install pinned tools
mise lock                # refresh lockfile after editing mise.toml
mise up                  # bump versions
```

## prek

- https://prek.j178.dev/ (pre-commit alternative)
- Hooks in `.pre-commit-config.yaml`

```sh
# Install git hooks:
prek install
# Update hook revs:
prek auto-update --cooldown-days 30

prek #same as prek run
prek run --all-files
```

## Tests

```sh
go test .
go test -run TestName .  # single test
go vet ./...
gofmt -w .
```

## Build

```sh
# Compile binary to project dir:
go build -o mud .
./mud -n "My File (2024).txt"
```

## Install locally

```sh
# Compile and install to $GOPATH/bin:
go install .
mud -n "My File (2024).txt"
# Clean:
go clean -i -x .
```

## Release

```sh
git tag v1.0.0
git push --tags
```

Kick the module proxy to index the new version immediately:

```sh
curl "https://proxy.golang.org/github.com/i15g/mud/@v/v1.0.0.info"
```
