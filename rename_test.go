package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	if err := runRename(src, renameOpts{}); err != nil {
		t.Fatal(err)
	}
	// Source should still exist (rename skipped)
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
