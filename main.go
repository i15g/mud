package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/spf13/pflag"
)

var version = func() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}()

func run(args []string, stdin io.Reader, isTTY bool) int {
	fs := pflag.NewFlagSet("mud", pflag.ContinueOnError)
	fs.SortFlags = false

	var (
		dryRun      bool
		quiet       bool
		verbose     bool
		interactive bool
		force       bool
		recursive   bool
		help        bool
		showVersion bool
	)

	fs.BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be renamed without renaming")
	fs.BoolVarP(&quiet, "quiet", "q", false, "Suppress all output")
	fs.BoolVarP(&verbose, "verbose", "v", false, "Print old -> new for each rename")
	fs.BoolVarP(&interactive, "interactive", "i", false, "Prompt before each rename")
	fs.BoolVarP(&force, "force", "f", false, "Bypass ignore patterns")
	fs.BoolVarP(&recursive, "recursive", "r", false, "Rename all files/dirs under path (bottom-up)")
	fs.BoolVarP(&help, "help", "h", false, "Show help")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 1
	}

	if showVersion {
		fmt.Fprintf(stderr, "mud %s\n", version)
		return 0
	}

	if help {
		printUsage(fs)
		return 0
	}

	opts := renameOpts{
		dryRun:      dryRun,
		quiet:       quiet,
		verbose:     verbose,
		interactive: interactive,
		force:       force,
	}

	positional := fs.Args()

	// Interactive requires a TTY
	if interactive && !isTTY {
		fmt.Fprintln(stderr, "mud: -i requires an interactive terminal")
		return 1
	}

	// Recursive mode
	if recursive {
		target := ""
		if len(positional) > 0 {
			target = positional[0]
		}
		if err := runRecursive(target, opts); err != nil {
			fmt.Fprintf(stderr, "mud: %s\n", err)
			return 1
		}
		return 0
	}

	// Stdin mode: not a TTY, no positional args
	if !isTTY && len(positional) == 0 {
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "mud: %s\n", err)
			return 1
		}
		input := string(data)
		// Trim single trailing newline if present (shell echo adds one)
		if len(input) > 0 && input[len(input)-1] == '\n' {
			input = input[:len(input)-1]
		}
		fmt.Fprintln(stdout, Sanitize(input))
		return 0
	}

	// No args
	if len(positional) == 0 {
		printUsage(fs)
		return 1
	}

	// Multi-arg iteration
	exitCode := 0
	for _, arg := range positional {
		if err := runRename(arg, opts); err != nil {
			if errors.Is(err, errQuit) {
				return 0 // user quit — exit zero
			}
			fmt.Fprintf(stderr, "mud: %s\n", err)
			exitCode = 1
		}
	}
	return exitCode
}

func printUsage(fs *pflag.FlagSet) {
	fmt.Fprintf(stderr, "Usage: mud %s [flags] <file>...\n\nRename files to URL-friendly format.\n\nFlags:\n", version)
	fs.PrintDefaults()
	fmt.Fprint(stderr, `
Examples:
  mud "My File.txt"           Rename to my-file.txt
  mud -n FOO.txt BAR.txt      Preview renames (dry run)
  mud -r .                    Rename all files recursively
  echo "Hello World" | mud    Sanitize text from stdin
`)
}

func main() {
	fi, _ := os.Stdin.Stat()
	isTTY := fi.Mode()&os.ModeCharDevice != 0
	os.Exit(run(os.Args[1:], os.Stdin, isTTY))
}
