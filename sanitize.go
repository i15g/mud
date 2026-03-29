package main

import "strings"

// isAlphanumeric returns true if s contains only [a-zA-Z0-9] characters.
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

// extractExtensions splits name into stem and up to 2 trailing dot-extensions.
// A dot-segment qualifies only if it consists entirely of [a-zA-Z0-9].
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

// Sanitize converts a filename (or text) to a URL-friendly format:
//   - Lowercase
//   - Separators (spaces, parens, brackets, etc.) → hyphens
//   - 3+ consecutive hyphens → 2 hyphens
//   - Non-[a-z0-9-] characters removed
//   - Leading/trailing hyphens removed
//   - Leading dots, leading underscores, and up to 2 trailing extensions preserved
func Sanitize(name string) string {
	// 1. Strip leading dots
	var leadingDots string
	for strings.HasPrefix(name, ".") {
		leadingDots += "."
		name = name[1:]
	}

	// 2. Extract up to 2 trailing extensions (only alphanumeric segments)
	var ext string
	name, ext = extractExtensions(name)

	// 3. Strip leading underscores from stem
	var leadingUnderscores string
	for strings.HasPrefix(name, "_") {
		leadingUnderscores += "_"
		name = name[1:]
	}

	// 4. Lowercase
	name = strings.ToLower(name)

	// 5. Replace separators with hyphens
	name = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '_', ',', '+', '.', '\u2014', '\u2013',
			';', ':', '(', ')', '{', '}', '[', ']', '|':
			return '-'
		}
		return r
	}, name)

	// 6. Collapse 3+ consecutive hyphens to 2
	name = collapseHyphens(name)

	// 7. Remove chars not in [a-z0-9-]
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1 // drop
	}, name)

	// 8. Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	// 9. Lowercase the extension
	ext = strings.ToLower(ext)

	return leadingDots + leadingUnderscores + name + ext
}

// collapseHyphens replaces 3 or more consecutive hyphens with 2.
func collapseHyphens(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	hyphenCount := 0
	for _, r := range s {
		if r == '-' {
			hyphenCount++
		} else {
			if hyphenCount >= 3 {
				b.WriteString("--")
			} else {
				for range hyphenCount {
					b.WriteByte('-')
				}
			}
			hyphenCount = 0
			b.WriteRune(r)
		}
	}
	// flush trailing hyphens
	if hyphenCount >= 3 {
		b.WriteString("--")
	} else {
		for range hyphenCount {
			b.WriteByte('-')
		}
	}
	return b.String()
}
