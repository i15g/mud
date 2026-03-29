package main

import (
	"fmt"
	"os"
)

const usageText = `Usage: mud [options] <filename|text>

Rename file to URL-friendly format:
  - Convert to lowercase
  - Replace spaces, underscores, commas, plus with hyphens
  - Remove non-alphanumeric characters (except hyphens and periods)
  - Collapse 3+ hyphens to 2
  - Preserve periods and leading underscore prefix

Options:
  -h, --help        Show this help
  -d, --dry-run     Sanitize filename and show what would be renamed
  -t, --text        Sanitize text and output result (don't rename file)
  -q, --quiet       Rename file and output only the new path
  -r, --recursive   Rename all files/dirs under path (bottom-up)

Examples:
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
`

type options struct {
	dryRun    bool
	textMode  bool
	quietMode bool
	recursive bool
	help      bool
	args      []string
}

func parseArgs(argv []string) (options, error) {
	var opts options
	i := 0
	for i < len(argv) {
		arg := argv[i]
		if arg == "--" {
			opts.args = append(opts.args, argv[i+1:]...)
			break
		}
		if len(arg) > 1 && arg[0] == '-' && arg[1] != '-' {
			// Short flag(s): -d, -rd, etc.
			for _, ch := range arg[1:] {
				switch ch {
				case 'h':
					opts.help = true
				case 'd':
					opts.dryRun = true
				case 't':
					opts.textMode = true
				case 'q':
					opts.quietMode = true
				case 'r':
					opts.recursive = true
				default:
					return opts, fmt.Errorf("unknown flag: -%c", ch)
				}
			}
		} else if len(arg) > 2 && arg[:2] == "--" {
			switch arg {
			case "--help":
				opts.help = true
			case "--dry-run":
				opts.dryRun = true
			case "--text":
				opts.textMode = true
			case "--quiet":
				opts.quietMode = true
			case "--recursive":
				opts.recursive = true
			default:
				return opts, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// First non-flag argument — rest are positional
			opts.args = append(opts.args, argv[i:]...)
			break
		}
		i++
	}
	return opts, nil
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.help {
		fmt.Print(usageText)
		os.Exit(0)
	}

	if opts.recursive {
		target := ""
		if len(opts.args) > 0 {
			target = opts.args[0]
		}
		rOpts := renameOpts{dryRun: opts.dryRun, quiet: opts.quietMode}
		if err := runRecursive(target, rOpts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if len(opts.args) == 0 {
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(1)
	}

	input := opts.args[0]

	if opts.textMode {
		fmt.Println(Sanitize(input))
		return
	}

	rOpts := renameOpts{dryRun: opts.dryRun, quiet: opts.quietMode}
	if err := runRename(input, rOpts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
