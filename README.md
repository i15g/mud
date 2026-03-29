# mud

Rename files to URL slug style format.

## Install

```sh
go install github.com/i15g/mud@latest
```

Or in a Brewfile:

```
go "github.com/i15g/mud"
```

## Usage

```
mud [flags] <file|directory|text>
```

### Flags

- `-n, --dry-run` - Show what would be renamed without actually renaming
- `-v, --verbose` - Show old → new filenames
- `-q, --quiet` - Suppress output
- `-f, --force` - Rename even if ignored (dotfiles, known files, etc.)
- `-i, --interactive` - Prompt before each rename (y/n/q)
- `-r, --recursive` - Rename files in subdirectories

### Examples

Rename a single file (interactive mode):

```sh
mud -i README.md
```

Dry-run on a directory:

```sh
mud -n -r /path/to/dir
```

Pipe text to stdin:

```sh
echo "Hello World" | mud
# Output: hello-world
```

Rename with verbose output:

```sh
mud -v -r .
```

Rename all files, ignoring errors (continue on clobber):

```sh
mud -r ./docs
```

## License

MIT
