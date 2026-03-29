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
