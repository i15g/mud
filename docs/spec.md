Usage: mud [options] <filename|text>

- Convert to lowercase
- Replace spaces, underscores, commas, plus with hyphens
  - TODO: other common separator chars?
- Remove non-alphanumeric characters (except hyphens and periods)
- Collapse 3+ hyphens to 2

TODO:
- Error if filename already exists (currently just silent noop?)

Options:
-h, --help        Show this help
-d, --dry-run     Sanitize filename and show what would be renamed
-t, --text        Sanitize text and output result (don't rename file)
-q, --quiet       Rename file and output only the new path
-r, --recursive   Rename all files/dirs under path (bottom-up)

Options TODO:
- Change `-d` to `-n` (match unix convention)
- quiet/verbose:
  - default: print just new filename (good default for scripting)
  - q/quiet: print nothing
  - v/verbose: print old -> new
- Interactive mode flag
  - mud -i FOO --> "rename FOO to foo?"
- config file
  - set default params
  - set custom ignore patterns
- Support multiple file args and globbing
  - mud FOO.txt BAR.txt
  - mud *.txt
- stdin support via isatty detection (like jq).
  - file: mud FOO.txt -> foo.txt
  - text: echo FOO | mud --> foo (can drop -t/--text flag, which had an ambiguous name anyway)
- possibly rm `r/recursive` flag, recommend fd + mud instead?
  - leverages fd's ignore behavior

Examples:
```
FOO                  -->  foo
foo bar              -->  foo-bar
foo_bar              -->  foo-bar
.bashrc              -->  .bashrc
.foo bar             -->  .foo-bar
_private file        -->  _private-file
My File (2024).txt   -->  my-file-2024.txt
foo.bar.baz          -->  foo.bar.baz
foo---bar            -->  foo--bar
hello world 2024     -->  hello-world-2024
```

## Allowed prefixes
- periods and underscores

## Ignore Patterns

These may be overridden with the -f/--force flag.

- dotfiles
- ALPHA.md (e.g. README.md)
- File extensions:
  - java, scala
  - python?
  - etc
- Known filenames:
  - Thumbs.db
  - desktop.ini
  - ehthumbs.db
