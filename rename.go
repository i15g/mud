package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

var errQuit = errors.New("quit")

type renameOpts struct {
	dryRun      bool
	quiet       bool
	verbose     bool
	interactive bool
	force       bool
	input       io.Reader // for interactive prompt; nil = os.Stdin
}

// runRename sanitizes the basename of input and renames it in-place.
// Returns an error if the rename fails or if the target would clobber an existing file.
func runRename(input string, opts renameOpts) error {
	dir := filepath.Dir(input)
	base := filepath.Base(input)

	// Ignore check (before sanitize)
	if !opts.force && ShouldIgnore(base) {
		return nil
	}

	sanitized := Sanitize(base)

	var output string
	if dir == "." {
		output = sanitized
	} else {
		output = filepath.Join(dir, sanitized)
	}

	// Already clean
	if input == output {
		if !opts.quiet && !opts.verbose {
			fmt.Fprintln(stdout, output)
		}
		return nil
	}

	// Clobber check
	if _, err := os.Lstat(output); err == nil {
		if !sameFile(input, output) {
			return fmt.Errorf("mud: %s: target already exists: %s", input, output)
		}
	}

	// Dry run
	if opts.dryRun {
		fmt.Fprintf(stdout, "%s -> %s (dry run)\n", input, output)
		return nil
	}

	// Interactive prompt
	if opts.interactive {
		r := opts.input
		if r == nil {
			r = os.Stdin
		}
		action, err := promptRename(input, output, r)
		if err != nil {
			return err
		}
		switch action {
		case 'n':
			return nil
		case 'q':
			return errQuit
		}
	}

	// Rename
	if err := os.Rename(input, output); err != nil {
		return err
	}

	// Output
	if opts.verbose {
		fmt.Fprintf(stdout, "%s -> %s\n", input, output)
	} else if !opts.quiet {
		fmt.Fprintln(stdout, output)
	}
	return nil
}

// sameFile reports whether paths a and b refer to the same filesystem object.
func sameFile(a, b string) bool {
	ia, err := os.Lstat(a)
	if err != nil {
		return false
	}
	ib, err := os.Lstat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ia, ib)
}

// promptRename prompts the user for confirmation of a rename.
// Returns 'y', 'n', or 'q' for yes, no, or quit.
func promptRename(oldName, newName string, r io.Reader) (byte, error) {
	fmt.Fprintf(stderr, "rename %s → %s? [y/n/q] ", oldName, newName)
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}
		switch buf[0] {
		case 'y', 'Y':
			return 'y', nil
		case 'n', 'N':
			return 'n', nil
		case 'q', 'Q':
			return 'q', nil
		case '\n', '\r':
			continue
		default:
			fmt.Fprintf(stderr, "rename %s → %s? [y/n/q] ", oldName, newName)
		}
	}
}

// runRecursive renames all files and directories under target (default "."),
// processing children before parents (bottom-up) so parent renames don't
// invalidate child paths. Errors during individual file renames are logged to stderr
// and don't stop processing, but the first error is returned at the end.
func runRecursive(target string, opts renameOpts) error {
	if target == "" {
		target = "."
	}

	var paths []string
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == target {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(paths)))

	var firstErr error
	for _, p := range paths {
		if err := runRename(p, opts); err != nil {
			if errors.Is(err, errQuit) {
				return nil // user quit — not an error
			}
			fmt.Fprintf(stderr, "mud: %s\n", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
