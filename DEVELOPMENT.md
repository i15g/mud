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
go test ./...
```

## Build

```sh
go build -o mud .
./mud -t "My File (2024).txt"
```

## Install locally

```sh
go install .
mud -t "My File (2024).txt"
```
