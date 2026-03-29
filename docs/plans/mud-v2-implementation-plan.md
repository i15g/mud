# mud v2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement all v2 features: expanded sanitize pipeline (URL-decode, accents, multi-extension, special replacements, new separators), ignore patterns, renameOpts refactor, pflag migration, stdin detection, multi-arg, interactive mode.

**Architecture:** Same 4-file `package main` structure. `Sanitize()` is a pure pipeline function. `ShouldIgnore()` is a new pure predicate. `runRename`/`runRecursive` take a `renameOpts` struct. `main.go` uses `github.com/spf13/pflag` for flag parsing and an extracted `run()` function for testability.

**Tech Stack:** Go 1.26, `github.com/spf13/pflag`, `net/url` (stdlib)

**Spec:** `docs/specs/v2.md` is the authoritative reference for all behavior.

---

## File Map

| File               | Changes                                                                                                                                                         |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `sanitize.go`      | Rewrite pipeline: URL-decode, accent map, multi-extension extraction, special replacements, expanded separator list, updated character filter                   |
| `sanitize_test.go` | Update 4 existing test expectations, add ~25 new cases                                                                                                          |
| `ignore.go`        | New file: `ShouldIgnore(name string) bool` with 4 predicates                                                                                                    |
| `ignore_test.go`   | New file: table-driven tests for all predicates                                                                                                                 |
| `rename.go`        | `renameOpts` struct, output modes, clobber error, dry-run format, interactive prompt, ShouldIgnore integration, package-level `stdout`/`stderr` for testability |
| `rename_test.go`   | Update 7 existing tests for new signatures, add ~15 new tests                                                                                                   |
| `main.go`          | pflag migration, `run()` extraction, stdin detection, multi-arg loop, help text                                                                                 |
| `go.mod`           | Add `github.com/spf13/pflag` dependency                                                                                                                         |

---

**Files:**

- Modify: `sanitize.go`
- Modify: `sanitize_test.go`

- [x] **Step 1: Write failing tests for multi-extension behavior**

Add these cases to the `tests` slice in `TestSanitize`:

```go
// Multi-extension (v2)
{"foo.tar.gz", "foo.tar.gz"},
{"My.Config.File.txt", "my-config.file.txt"},
{"hello world.foo bar baz.txt", "hello-world-foo-bar-baz.txt"},
{"archive.tar.GZ", "archive.tar.gz"},
{"_.txt", "_.txt"},
```

Note: `My.Config.File.txt` produces `my-config.file.txt` because 2 qualifying extensions (`.File.txt`) are extracted, leaving stem `My.Config`. Period in stem becomes hyphen at step 7 (Task 2). This test will only fully pass after Task 2 — for now it validates the extension extraction.

Expected interim behavior after this task only: `My.Config.File.txt` → `my.config.file.txt` (period in stem preserved since step 7 isn't expanded yet). We'll update the expected value to `my-config.file.txt` in Task 2.

Actually, to keep tests green at each commit, use the interim expectation:

```go
{"My.Config.File.txt", "my.config.file.txt"}, // updated to my-config.file.txt in Task 2
```

- [x] **Step 2: Run tests to verify they fail**

Run: `go test -run TestSanitize -v ./...`
Expected: `foo.tar.gz` may already pass (coincidence — v1 extracts `.gz`, stem `foo.tar`, period preserved). `My.Config.File.txt` will fail. `hello world.foo bar baz.txt` will fail.

- [x] **Step 3: Implement `extractExtensions` helper and `isAlphanumeric` helper**

Add to `sanitize.go`:

```go
// extractExtensions splits name into stem and up to 2 trailing dot-extensions.
// A dot-segment qualifies only if it consists entirely of [a-zA-Z0-9].
func extractExtensions(name string) (string, string) {
	remaining := name
	var extensions string
	for range 2 {
		i := strings.LastIndex(remaining, ".")
		if i < 0 {
			break
		}
		segment := remaining[i+1:]
		if !isAlphanumeric(segment) {
			break
		}
		extensions = remaining[i:] + extensions[len(remaining[i+1:])+1:]
		// Simpler: rebuild
		extensions = "." + segment + extensions[0:0]
		// Actually let me just do this correctly:
		remaining = remaining[:i]
	}
	// Hmm, let me write this more carefully
	return remaining, extensions
}
```

Actually, cleaner implementation:

```go
func extractExtensions(name string) (stem, ext string) {
	remaining := name
	collected := ""
	for range 2 {
		i := strings.LastIndex(remaining, ".")
		if i < 0 {
			break
		}
		seg := remaining[i+1:]
		if !isAlphanumeric(seg) {
			break
		}
		collected = remaining[i:] + collected
		remaining = remaining[:i]
	}
	return remaining, collected
}

func isAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
```

Wait, there's a bug in the collected building. Let me trace:

For `foo.tar.gz`:

- Iter 1: remaining=`foo.tar.gz`, i=7, seg=`gz`, isAlpha=true. collected = `foo.tar.gz`[7:] + "" = `.gz`. remaining=`foo.tar`.
- Iter 2: remaining=`foo.tar`, i=3, seg=`tar`, isAlpha=true. collected = `foo.tar`[3:] + `.gz` = `.tar` + `.gz` = `.tar.gz`. remaining=`foo`.
- Return: `foo`, `.tar.gz` ✓

For `My.Config.File.txt`:

- Iter 1: remaining=`My.Config.File.txt`, i=14, seg=`txt`. collected=`.txt`. remaining=`My.Config.File`.
- Iter 2: remaining=`My.Config.File`, i=9, seg=`File`. collected=`.File` + `.txt` = `.File.txt`. remaining=`My.Config`.
- Return: `My.Config`, `.File.txt` ✓

For `hello world.foo bar baz.txt`:

- Iter 1: remaining=full, i at last `.`, seg=`txt`. collected=`.txt`. remaining=`hello world.foo bar baz`.
- Iter 2: remaining=`hello world.foo bar baz`, last `.` at 11, seg=`foo bar baz`. isAlpha=false (spaces). Break.
- Return: `hello world.foo bar baz`, `.txt` ✓

The `collected = remaining[i:] + collected` pattern works because `remaining[i:]` is `.seg` and `collected` already has the previously collected extensions. Let me verify once more:

Iter 1: collected = "" initially. `remaining[i:]` = `.gz`. collected = `.gz` + `""` = `.gz`.
Iter 2: `remaining[i:]` = `.tar`. collected = `.tar` + `.gz` = `.tar.gz`.

Yes, this is correct. The final implementation:

```go
func extractExtensions(name string) (string, string) {
	remaining := name
	collected := ""
	for range 2 {
		i := strings.LastIndex(remaining, ".")
		if i < 0 {
			break
		}
		if !isAlphanumeric(remaining[i+1:]) {
			break
		}
		collected = remaining[i:] + collected
		remaining = remaining[:i]
	}
	return remaining, collected
}

func isAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
```

- [x] **Step 4: Replace extension extraction in `Sanitize` with `extractExtensions`**

In `sanitize.go`, replace lines 20-25:

```go
// Old:
var ext string
if i := strings.LastIndex(name, "."); i >= 0 {
    ext = name[i:]
    name = name[:i]
}

// New:
name, ext := extractExtensions(name)
```

Note: `ext` was previously declared with `var`; the new code uses `:=` in the multi-return. Adjust variable declarations in the function accordingly — `name` is a parameter so it can be reassigned, but `ext` needs to be introduced via `:=` or declared earlier. Since `name` is reassigned by `extractExtensions`, the cleanest approach:

```go
var ext string
name, ext = extractExtensions(name)
```

- [x] **Step 5: Run tests to verify they pass**

Run: `go test -run TestSanitize -v ./...`
Expected: All tests pass (including existing ones — multi-extension extraction produces the same results for all existing test inputs).

- [x] **Step 6: Commit**

```bash
git add sanitize.go sanitize_test.go
git commit -m "feat: multi-extension extraction in Sanitize pipeline"
```

---

**Files:**

- Modify: `sanitize.go`
- Modify: `sanitize_test.go`

- [x] **Step 1: Update existing test expectations that change**

In `sanitize_test.go`, update these cases:

```go
// Was: {"foo(bar)", "foobar"}
{"foo(bar)", "foo-bar"},

// Was: {"My File (2024).txt", "my-file-2024.txt"}
{"My File (2024).txt", "my-file--2024.txt"},
```

And update the interim multi-extension case from Task 1:

```go
// Was: {"My.Config.File.txt", "my.config.file.txt"}
{"My.Config.File.txt", "my-config.file.txt"},
```

- [x] **Step 2: Add new test cases for expanded separators**

```go
// New separators (v2)
{"foo(bar).txt", "foo-bar.txt"},
{"foo[1].txt", "foo-1.txt"},
{"a;b", "a-b"},
{"a:b", "a-b"},
{"a|b", "a-b"},
{"a\u2014b", "a-b"}, // em dash
{"a\u2013b", "a-b"}, // en dash
{"a{b}", "a-b"},
```

- [x] **Step 3: Run tests to verify they fail**

Run: `go test -run TestSanitize -v ./...`
Expected: New separator tests fail (parens, brackets, etc. currently stripped, not converted to hyphens). Updated expectations for `foo(bar)` and `My File (2024).txt` also fail.

- [x] **Step 4: Expand separator list and update character filter**

In `sanitize.go`, update the separator replacement (step 5/step 7 in the pipeline):

```go
// Replace separators with hyphens
name = strings.Map(func(r rune) rune {
    switch r {
    case ' ', '\n', '_', ',', '+', '.', '\u2014', '\u2013',
        ';', ':', '(', ')', '{', '}', '[', ']', '|':
        return '-'
    }
    return r
}, name)
```

And update the character filter to disallow periods in stem:

```go
// Remove chars not in [a-z0-9-]
name = strings.Map(func(r rune) rune {
    if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
        return r
    }
    return -1
}, name)
```

Note the removal of `|| r == '.'` from the filter — periods in stem are now converted to hyphens at step 7, and any remaining periods are stripped.

- [x] **Step 5: Run tests to verify they pass**

Run: `go test -run TestSanitize -v ./...`
Expected: All tests pass.

- [x] **Step 6: Commit**

```bash
git add sanitize.go sanitize_test.go
git commit -m "feat: expanded separator list and period-in-stem handling"
```

---

**Files:**

- Modify: `sanitize.go`
- Modify: `sanitize_test.go`

- [x] **Step 1: Update existing accent test expectations**

```go
// Was: {"café", "caf"}
{"café", "cafe"},

// Was: {"naïve", "nave"}
{"naïve", "naive"},
```

- [x] **Step 2: Add new test cases for URL decoding, accents, and special replacements**

```go
// URL decoding (v2 step 0)
{"hello%20world.txt", "hello-world.txt"},
{"foo%28bar%29.txt", "foo-bar.txt"},

// Accent normalization (v2 step 1)
{"\u00c9tude.txt", "etude.txt"},     // É → e
{"stra\u00dfe", "strasse"},           // ß → ss
{"pi\u00f1ata", "pinata"},            // ñ → n
{"fa\u00e7ade", "facade"},            // ç → c
{"\u00e0 la carte", "a-la-carte"},    // à → a

// Special replacements (v2 step 6)
{"foo@bar", "foo-at-bar"},
{"a&b", "a-and-b"},
{"@foo.txt", "at-foo.txt"},
```

- [x] **Step 3: Run tests to verify they fail**

Run: `go test -run TestSanitize -v ./...`
Expected: All new tests and updated accent tests fail.

- [x] **Step 4: Implement accent normalization map**

Add to `sanitize.go` at package level:

```go
var accentMap = map[rune]string{
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a",
	'À': "a", 'Á': "a", 'Â': "a", 'Ã': "a", 'Ä': "a", 'Å': "a",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'È': "e", 'É': "e", 'Ê': "e", 'Ë': "e",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'Ì': "i", 'Í': "i", 'Î': "i", 'Ï': "i",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o",
	'Ò': "o", 'Ó': "o", 'Ô': "o", 'Õ': "o", 'Ö': "o",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
	'Ù': "u", 'Ú': "u", 'Û': "u", 'Ü': "u",
	'ñ': "n", 'Ñ': "n",
	'ç': "c", 'Ç': "c",
	'ß': "ss",
}
```

Add a `normalizeAccents` helper:

```go
func normalizeAccents(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if repl, ok := accentMap[r]; ok {
			b.WriteString(repl)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
```

- [x] **Step 5: Implement URL decoding and special replacements; wire all new steps into `Sanitize`**

Add import `"net/url"` to `sanitize.go`.

Add a `specialReplacements` helper:

```go
func specialReplacements(s string) string {
	s = strings.ReplaceAll(s, "@", "-at-")
	s = strings.ReplaceAll(s, "&", "-and-")
	return s
}
```

Update `Sanitize` to add the new steps in order. The full pipeline becomes:

```go
func Sanitize(name string) string {
	// 0. URL-decode
	if decoded, err := url.PathUnescape(name); err == nil {
		name = decoded
	}

	// 1. Accent normalization
	name = normalizeAccents(name)

	// 2. Strip leading dots
	var leadingDots string
	for strings.HasPrefix(name, ".") {
		leadingDots += "."
		name = name[1:]
	}

	// 3. Extract up to 2 trailing extensions
	var ext string
	name, ext = extractExtensions(name)

	// 4. Strip leading underscores from stem
	var leadingUnderscores string
	for strings.HasPrefix(name, "_") {
		leadingUnderscores += "_"
		name = name[1:]
	}

	// 5. Lowercase stem
	name = strings.ToLower(name)

	// 6. Special replacements
	name = specialReplacements(name)

	// 7. Replace separators with hyphens
	name = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '_', ',', '+', '.', '\u2014', '\u2013',
			';', ':', '(', ')', '{', '}', '[', ']', '|':
			return '-'
		}
		return r
	}, name)

	// 8. Collapse 3+ consecutive hyphens to 2
	name = collapseHyphens(name)

	// 9. Remove chars not in [a-z0-9-]
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, name)

	// 10. Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	// 11. Lowercase extension
	ext = strings.ToLower(ext)

	return leadingDots + leadingUnderscores + name + ext
}
```

- [x] **Step 6: Run tests to verify they pass**

Run: `go test -run TestSanitize -v ./...`
Expected: All tests pass.

- [x] **Step 7: Commit**

```bash
git add sanitize.go sanitize_test.go
git commit -m "feat: URL decoding, accent normalization, and special replacements"
```

---

**Files:**

- Create: `ignore.go`
- Create: `ignore_test.go`

- [x] **Step 1: Write `TestShouldIgnore` with table-driven tests**

Create `ignore_test.go`:

```go
package main

import "testing"

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		// Predicate 1: Dotfiles
		{".bashrc", true},
		{".gitignore", true},
		{"..hidden", true},
		{".DS_Store", true},
		{"._foo", true},

		// Predicate 2: All-caps stems
		{"README", true},
		{"README.md", true},
		{"LICENSE", true},
		{"CONTRIBUTING", true},
		{"CHANGELOG.md", true},
		{"RELEASE-NOTES", false},  // hyphen in stem — not protected
		{"FOO123", false},         // digits — not ^[A-Z]+$
		{"Foo", false},            // mixed case
		{"foo", false},            // lowercase

		// Predicate 3: Known filenames
		{"Thumbs.db", true},
		{"desktop.ini", true},
		{"ehthumbs.db", true},
		{"Makefile", true},
		{"makefile", false},  // exact match only
		{"thumbs.db", false}, // exact match only

		// Predicate 4: Convention extensions
		{"main.go", true},
		{"script.py", true},
		{"lib.rs", true},
		{"Gemfile.rb", true},
		{"hello.c", true},
		{"hello.h", true},
		{"hello.cpp", true},
		{"hello.hpp", true},
		{"mix.ex", true},
		{"mix.exs", true},
		{"gen_server.erl", true},
		{"MyClass.java", true},
		{"Main.scala", true},
		{"Main.kt", true},
		{"Program.cs", true},
		{"App.swift", true},

		// Not ignored
		{"foo.txt", false},
		{"foo.md", false},
		{"foo.html", false},
		{"My File.txt", false},
		{"photo.jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldIgnore(tt.name)
			if got != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
```

- [x] **Step 2: Run tests to verify they fail**

Run: `go test -run TestShouldIgnore -v ./...`
Expected: Compilation error — `ShouldIgnore` not defined.

- [x] **Step 3: Implement `ShouldIgnore` in `ignore.go`**

Create `ignore.go`:

```go
package main

import (
	"path/filepath"
	"strings"
)

var knownFiles = map[string]bool{
	"Thumbs.db":   true,
	"desktop.ini": true,
	"ehthumbs.db": true,
	"Makefile":    true,
}

var conventionExts = map[string]bool{
	// snake_case languages
	".go": true, ".py": true, ".rs": true, ".rb": true,
	".c": true, ".h": true, ".cpp": true, ".hpp": true,
	".ex": true, ".exs": true, ".erl": true,
	// PascalCase languages
	".java": true, ".scala": true, ".kt": true, ".cs": true, ".swift": true,
}

// ShouldIgnore reports whether name should be skipped during rename.
// All predicates are bypassed by the -f/--force flag (handled by caller).
func ShouldIgnore(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	if allCaps(strings.TrimSuffix(name, filepath.Ext(name))) {
		return true
	}
	if knownFiles[name] {
		return true
	}
	if conventionExts[filepath.Ext(name)] {
		return true
	}
	return false
}

// allCaps reports whether s is non-empty and consists entirely of A-Z.
func allCaps(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
```

- [x] **Step 4: Run tests to verify they pass**

Run: `go test -run TestShouldIgnore -v ./...`
Expected: All pass.

- [x] **Step 5: Commit**

```bash
git add ignore.go ignore_test.go
git commit -m "feat: ShouldIgnore predicate for ignore patterns"
```

---

**Files:**

- Modify: `rename.go`
- Modify: `rename_test.go`
- Modify: `main.go` (call-site updates only)

This task defines the `renameOpts` struct, introduces package-level `stdout`/`stderr` writers for testability, updates all existing test call sites, and updates `main.go` call sites to bridge old flags to new struct. No behavior changes yet — purely mechanical refactor.

- [x] **Step 1: Add `renameOpts` struct and output writers to `rename.go`**

Add at the top of `rename.go`, after imports:

```go
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
```

Add `"io"` to the import block.

- [x] **Step 2: Update `runRename` signature and call sites**

Change `runRename` signature from:

```go
func runRename(input string, dryRun, quiet bool) error {
```

to:

```go
func runRename(input string, opts renameOpts) error {
```

Inside the function body, replace `dryRun` with `opts.dryRun` and `quiet` with `opts.quiet`. Replace `fmt.Printf` / `fmt.Println` with `fmt.Fprintf(stdout, ...)` / `fmt.Fprintln(stdout, ...)`. Keep v1 output behavior for now (changed in Task 6).

- [x] **Step 3: Update `runRecursive` signature**

Change from:

```go
func runRecursive(target string, dryRun, quiet bool) error {
```

to:

```go
func runRecursive(target string, opts renameOpts) error {
```

Update the `runRename` call inside to pass `opts`.

- [x] **Step 4: Update `main.go` call sites**

Bridge the old flag values into the new struct:

```go
// In the recursive branch:
opts := renameOpts{dryRun: opts.dryRun, quiet: opts.quietMode}
if err := runRecursive(target, opts); err != nil {

// In the default rename branch:
rOpts := renameOpts{dryRun: opts.dryRun, quiet: opts.quietMode}
if err := runRename(input, rOpts); err != nil {
```

Note: avoid variable shadowing — use `rOpts` for `renameOpts` to distinguish from the CLI `options` struct.

- [x] **Step 5: Update all existing tests in `rename_test.go`**

Mechanical replacement:

```go
// TestRunRename_Basic:
runRename(src, renameOpts{})

// TestRunRename_AlreadyClean:
runRename(src, renameOpts{})

// TestRunRename_NoClobber:
runRename(src, renameOpts{})

// TestRunRename_DryRun:
runRename(src, renameOpts{dryRun: true})

// TestRunRename_Quiet:
runRename(src, renameOpts{quiet: true})

// TestRunRecursive_BottomUp:
runRecursive(dir, renameOpts{})

// TestRunRecursive_DryRun:
runRecursive(dir, renameOpts{dryRun: true})

// TestRunRecursive_SkipsGit:
runRecursive(dir, renameOpts{})
```

Add `"bytes"` to the import block for the output capture helper used in later tasks.

- [x] **Step 6: Run all tests**

Run: `go test -v ./...`
Expected: All tests pass (pure refactor, no behavior change).

- [x] **Step 7: Commit**

```bash
git add rename.go rename_test.go main.go
git commit -m "refactor: renameOpts struct and output writer infrastructure"
```

---

**Files:**

- Modify: `rename.go`
- Modify: `rename_test.go`

This task implements: clobber error, output mode changes (default/quiet/verbose), dry-run format change, already-clean behavior changes, ShouldIgnore integration, and continue-on-error in `runRecursive`.

- [x] **Step 1: Add output capture helper to `rename_test.go`**

```go
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := stdout
	stdout = &buf
	defer func() { stdout = old }()
	fn()
	return buf.String()
}
```

- [x] **Step 2: Write failing test for clobber error**

Update `TestRunRename_NoClobber`:

```go
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
```

- [x] **Step 3: Write failing tests for output modes**

```go
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
```

- [x] **Step 4: Write failing tests for dry-run format**

```go
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
```

- [x] **Step 5: Write failing tests for already-clean behavior**

```go
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
```

- [x] **Step 6: Write failing tests for ShouldIgnore integration**

```go
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
	assertExists(t, filepath.Join(dir, "readme.md"))
	assertNotExists(t, src)
}
```

- [x] **Step 7: Run tests to see failures**

Run: `go test -run TestRunRename -v ./...`
Expected: Multiple failures — clobber still returns nil, output modes wrong, already-clean doesn't print, no ignore integration.

- [x] **Step 8: Implement all behavior changes in `runRename`**

Rewrite `runRename` in `rename.go`:

```go
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
```

- [x] **Step 9: Update `runRecursive` to continue on error**

Add a sentinel for interactive quit (will be used in Task 7):

```go
var errQuit = errors.New("quit")
```

Add `"errors"` to the import block.

Update `runRecursive`:

```go
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
```

- [x] **Step 10: Run all tests**

Run: `go test -v ./...`
Expected: All tests pass.

- [x] **Step 11: Commit**

```bash
git add rename.go rename_test.go
git commit -m "feat: clobber error, output modes, dry-run format, ShouldIgnore integration"
```

---

**Files:**

- Modify: `rename.go`
- Modify: `rename_test.go`

- [x] **Step 1: Write failing tests for interactive mode**

```go
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
```

Add `"errors"` to the test file imports.

- [x] **Step 2: Run tests to verify they fail**

Run: `go test -run TestRunRename_Interactive -v ./...`
Expected: Fails — interactive prompt not implemented.

- [x] **Step 3: Implement `promptRename` function and integrate into `runRename`**

Add to `rename.go`:

```go
func promptRename(oldName, newName string, r io.Reader) (byte, error) {
	fmt.Fprintf(stderr, "rename %s \u2192 %s? [y/n/q] ", oldName, newName)
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
			fmt.Fprintf(stderr, "rename %s \u2192 %s? [y/n/q] ", oldName, newName)
		}
	}
}
```

In `runRename`, add the interactive prompt after the dry-run check and before the actual rename:

```go
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
```

- [x] **Step 4: Run tests to verify they pass**

Run: `go test -v ./...`
Expected: All tests pass.

- [x] **Step 5: Commit**

```bash
git add rename.go rename_test.go
git commit -m "feat: interactive prompt for rename confirmation"
```

---

**Files:**

- Modify: `main.go`
- Modify: `go.mod`
- Create: `go.sum` (auto-generated)

- [ ] **Step 1: Add pflag dependency**

```bash
go get github.com/spf13/pflag
```

- [ ] **Step 2: Rewrite `main.go` with pflag and `run()` extraction**

Replace the entire contents of `main.go`:

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/pflag"
)

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
	)

	fs.BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be renamed without renaming")
	fs.BoolVarP(&quiet, "quiet", "q", false, "Suppress all output")
	fs.BoolVarP(&verbose, "verbose", "v", false, "Print old -> new for each rename")
	fs.BoolVarP(&interactive, "interactive", "i", false, "Prompt before each rename")
	fs.BoolVarP(&force, "force", "f", false, "Bypass ignore patterns")
	fs.BoolVarP(&recursive, "recursive", "r", false, "Rename all files/dirs under path (bottom-up)")
	fs.BoolVarP(&help, "help", "h", false, "Show help")

	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 1
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
	fmt.Fprint(stderr, `Usage: mud [flags] <file>...

Rename files to URL-friendly format.

Flags:
`)
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
```

- [ ] **Step 3: Delete old `parseArgs` function and `options` struct**

The old `parseArgs`, `options` struct, and `usageText` constant are all replaced by the new code. Verify they are fully removed.

- [ ] **Step 4: Run all tests**

Run: `go test -v ./...`
Expected: All tests pass. The old `main.go` code is gone, but all rename/sanitize/ignore tests are independent of `main.go`.

- [ ] **Step 5: Run `go vet` and verify no issues**

```bash
go vet ./...
```

- [ ] **Step 6: Manual smoke test**

Build and test basic functionality:

```bash
go build -o mud .
echo "Hello World" | ./mud
# Expected: hello-world

./mud -h
# Expected: help text with pflag-formatted flags

./mud -n "FOO.txt"
# Expected: FOO.txt -> foo.txt (dry run)
```

(These are non-destructive commands — no files are modified.)

- [ ] **Step 7: Commit**

```bash
git add main.go go.mod go.sum
git commit -m "feat: pflag migration, stdin detection, multi-arg support"
```

---

**Files:**

- Modify: `rename_test.go` (add recursive + ignore integration test)
- Modify: `README.md`

- [x] **Step 1: Write integration test for recursive + ignore**

Add to `rename_test.go`:

```go
func TestRunRecursive_IgnoresConventionFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.go"), "package main")
	writeFile(t, filepath.Join(dir, "README.md"), "# title")
	writeFile(t, filepath.Join(dir, "My File.txt"), "content")

	if err := runRecursive(dir, renameOpts{}); err != nil {
		t.Fatal(err)
	}

	// Ignored files unchanged
	assertExists(t, filepath.Join(dir, "main.go"))
	assertExists(t, filepath.Join(dir, "README.md"))
	// Non-ignored file renamed
	assertExists(t, filepath.Join(dir, "my-file.txt"))
}

func TestRunRecursive_ForceOverridesIgnore(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "README.md"), "# title")

	if err := runRecursive(dir, renameOpts{force: true}); err != nil {
		t.Fatal(err)
	}

	assertExists(t, filepath.Join(dir, "readme.md"))
	assertNotExists(t, filepath.Join(dir, "README.md"))
}

func TestRunRecursive_ContinuesOnError(t *testing.T) {
	dir := t.TempDir()
	// Create two files that would produce the same sanitized name
	writeFile(t, filepath.Join(dir, "My File.txt"), "first")
	writeFile(t, filepath.Join(dir, "my-file.txt"), "existing")

	err := runRecursive(dir, renameOpts{})
	if err == nil {
		t.Fatal("expected error from clobber, got nil")
	}
}
```

- [x] **Step 2: Run tests**

Run: `go test -v ./...`
Expected: All pass.

- [x] **Step 3: Update README.md**

Update the usage examples and flag documentation to reflect v2 changes:

- Replace `-d` with `-n` for dry-run
- Remove `-t` / `--text`
- Add `-v`, `-i`, `-f` flags
- Update output mode descriptions
- Add stdin pipe example

- [x] **Step 4: Commit**

```bash
git add rename_test.go README.md
git commit -m "feat: integration tests and updated README for v2"
```

---

## Verification

After all tasks are complete, run the full verification suite:

```bash
# All tests pass
go test -v ./...

# No vet issues
go vet ./...

# Build succeeds
go build -o mud .

# Smoke tests
echo "Hello World" | ./mud                    # → hello-world
echo "café" | ./mud                           # → cafe
echo "foo%20bar" | ./mud                      # → foo-bar
./mud -n "My File (2024).txt"                 # → My File (2024).txt -> my-file--2024.txt (dry run)
./mud -h                                      # → help with pflag formatting

# Create temp files for rename tests
mkdir -p /tmp/mud-test && cd /tmp/mud-test
touch "FOO.txt" "BAR.txt" "README.md" "main.go"
/path/to/mud -v FOO.txt BAR.txt              # → renames FOO and BAR, skips README and main.go
/path/to/mud -vf README.md                   # → force-renames README.md
```
