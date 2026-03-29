package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := stdout
	stdout = &buf
	defer func() { stdout = old }()
	fn()
	return buf.String()
}

func TestRunRename_Basic(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.TXT")
	writeFile(t, src, "content")

	if err := runRename(src, renameOpts{}); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "my-file.txt")
	assertExists(t, dst)
	assertNotExists(t, src)
}

func TestRunRename_AlreadyClean(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "already-clean.txt")
	writeFile(t, src, "content")

	if err := runRename(src, renameOpts{}); err != nil {
		t.Fatal(err)
	}
	// File should still exist at original path
	assertExists(t, src)
}

func TestRunRename_NoClobber(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	dst := filepath.Join(dir, "my-file.txt")
	writeFile(t, src, "source")
	writeFile(t, dst, "existing")

	err := runRename(src, renameOpts{})
	if err == nil {
		t.Fatal("expected clobber error, got nil")
	}
	if !strings.Contains(err.Error(), "target already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
	// Source should still exist
	assertExists(t, src)
	// Destination content unchanged
	got := readFile(t, dst)
	if got != "existing" {
		t.Errorf("destination overwritten: got %q", got)
	}
}

func TestRunRename_DryRun(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	if err := runRename(src, renameOpts{dryRun: true}); err != nil {
		t.Fatal(err)
	}
	// Source must still exist after dry run
	assertExists(t, src)
	// Destination must not exist
	assertNotExists(t, filepath.Join(dir, "my-file.txt"))
}

func TestRunRename_Quiet(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	if err := runRename(src, renameOpts{quiet: true}); err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(dir, "my-file.txt"))
	assertNotExists(t, src)
}

func TestRunRename_DefaultOutput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{}); err != nil {
			t.Fatal(err)
		}
	})
	want := filepath.Join(dir, "my-file.txt") + "\n"
	if out != want {
		t.Errorf("default output = %q, want %q", out, want)
	}
}

func TestRunRename_VerboseOutput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{verbose: true}); err != nil {
			t.Fatal(err)
		}
	})
	want := src + " -> " + filepath.Join(dir, "my-file.txt") + "\n"
	if out != want {
		t.Errorf("verbose output = %q, want %q", out, want)
	}
}

func TestRunRename_QuietOutput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{quiet: true}); err != nil {
			t.Fatal(err)
		}
	})
	if out != "" {
		t.Errorf("quiet output = %q, want empty", out)
	}
}

func TestRunRename_DryRunFormat(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{dryRun: true}); err != nil {
			t.Fatal(err)
		}
	})
	want := src + " -> " + filepath.Join(dir, "my-file.txt") + " (dry run)\n"
	if out != want {
		t.Errorf("dry-run output = %q, want %q", out, want)
	}
}

func TestRunRename_AlreadyClean_Default(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "already-clean.txt")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{}); err != nil {
			t.Fatal(err)
		}
	})
	// Default mode prints the (unchanged) name
	want := src + "\n"
	if out != want {
		t.Errorf("already-clean default = %q, want %q", out, want)
	}
}

func TestRunRename_AlreadyClean_Verbose(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "already-clean.txt")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{verbose: true}); err != nil {
			t.Fatal(err)
		}
	})
	// Verbose skips silently for already-clean
	if out != "" {
		t.Errorf("already-clean verbose = %q, want empty", out)
	}
}

func TestRunRename_Ignored(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "README.md")
	writeFile(t, src, "content")

	out := captureOutput(t, func() {
		if err := runRename(src, renameOpts{}); err != nil {
			t.Fatal(err)
		}
	})
	assertExists(t, src)
	if out != "" {
		t.Errorf("ignored file should produce no output, got %q", out)
	}
}

func TestRunRename_Ignored_Force(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "README.md")
	writeFile(t, src, "content")

	if err := runRename(src, renameOpts{force: true}); err != nil {
		t.Fatal(err)
	}
	// After forcing, the file should be renamed (on case-insensitive filesystems,
	// README.md and readme.md refer to the same file after rename).
	// Verify the sanitized name exists.
	assertExists(t, filepath.Join(dir, "readme.md"))
}

func TestRunRecursive_BottomUp(t *testing.T) {
	// Build: root/My Dir/My File.txt
	// After recursive: root/my-dir/my-file.txt
	dir := t.TempDir()
	subdir := filepath.Join(dir, "My Dir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(subdir, "My File.txt"), "content")

	if err := runRecursive(dir, renameOpts{}); err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dir, "my-dir", "my-file.txt"))
	assertNotExists(t, filepath.Join(dir, "My Dir"))
}

func TestRunRecursive_DryRun(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "My Dir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(subdir, "My File.txt"), "content")

	if err := runRecursive(dir, renameOpts{dryRun: true}); err != nil {
		t.Fatal(err)
	}

	// Nothing should have changed
	assertExists(t, filepath.Join(dir, "My Dir", "My File.txt"))
	assertNotExists(t, filepath.Join(dir, "my-dir"))
}

func TestRunRecursive_SkipsGit(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(gitDir, "HEAD"), "ref: refs/heads/main")
	writeFile(t, filepath.Join(dir, "My File.txt"), "content")

	if err := runRecursive(dir, renameOpts{}); err != nil {
		t.Fatal(err)
	}

	// .git directory must be untouched
	assertExists(t, filepath.Join(dir, ".git", "HEAD"))
	// Regular file renamed
	assertExists(t, filepath.Join(dir, "my-file.txt"))
}

func TestRunRename_Interactive_Accept(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	opts := renameOpts{interactive: true, input: strings.NewReader("y\n")}
	if err := runRename(src, opts); err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(dir, "my-file.txt"))
	assertNotExists(t, src)
}

func TestRunRename_Interactive_Skip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	opts := renameOpts{interactive: true, input: strings.NewReader("n\n")}
	if err := runRename(src, opts); err != nil {
		t.Fatal(err)
	}
	assertExists(t, src)
	assertNotExists(t, filepath.Join(dir, "my-file.txt"))
}

func TestRunRename_Interactive_Quit(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "My File.txt")
	writeFile(t, src, "content")

	opts := renameOpts{interactive: true, input: strings.NewReader("q\n")}
	err := runRename(src, opts)
	if !errors.Is(err, errQuit) {
		t.Fatalf("expected errQuit, got %v", err)
	}
	assertExists(t, src)
}

// helpers

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimRight(string(b), "\n")
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", path)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err == nil {
		t.Errorf("expected %s to not exist", path)
	}
}
