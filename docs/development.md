# Development

## prek

```sh
# Install git hooks:
prek install
# Update hook revs:
prek auto-update --cooldown-days 30
```

## Tests

```sh
go test .
```

## Build

```sh
# Compile binary to project dir:
go build -o mud .
./mud -t "My File (2024).txt"
```

## Install locally

```sh
# Compile and install to $GOPATH/bin:
go install .
mud -t "My File (2024).txt"
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
