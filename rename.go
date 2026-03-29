package main

import (
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

type renameOpts struct {
	dryRun      bool
	quiet       bool
	verbose     bool
	interactive bool
	force       bool
	input       io.Reader // for interactive prompt; nil = os.Stdin
}

// runRename sanitizes the basename of input and renames it in-place.
// It silently skips if the name is already clean or the target already exists
// (as a different file). Returns an error only if the rename itself fails.
func runRename(input string, opts renameOpts) error {
	dir := filepath.Dir(input)
	base := filepath.Base(input)
	sanitized := Sanitize(base)

	var output string
	if dir == "." {
		output = sanitized
	} else {
		output = filepath.Join(dir, sanitized)
	}

	// Already clean — silent skip
	if input == output {
		return nil
	}

	// No-clobber check: if target exists and is a different file, skip silently.
	// Use os.SameFile to allow case-only renames (e.g. FOO → foo) on
	// case-insensitive filesystems where Lstat("foo") succeeds when "FOO" exists.
	if _, err := os.Lstat(output); err == nil {
		if !sameFile(input, output) {
			return nil
		}
	}

	if opts.dryRun {
		fmt.Fprintf(stdout, "%s --> %s\n(dry run)\n", input, output)
		return nil
	}

	if err := os.Rename(input, output); err != nil {
		return err
	}

	if opts.quiet {
		fmt.Fprintln(stdout, output)
	} else {
		fmt.Fprintf(stdout, "%s -> %s\n", input, output)
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

// runRecursive renames all files and directories under target (default "."),
// processing children before parents (bottom-up) so parent renames don't
// invalidate child paths.
func runRecursive(target string, opts renameOpts) error {
	if target == "" {
		target = "."
	}

	var paths []string
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip the root itself and .git directories
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

	// Sort reverse so deepest paths come first (bottom-up traversal)
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))

	for _, p := range paths {
		if err := runRename(p, opts); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}
