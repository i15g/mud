package main

import (
	"net/url"
	"strings"
)

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

// normalizeAccents replaces accented characters with their ASCII base equivalents.
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

// specialReplacements handles @ and & character replacements.
func specialReplacements(s string) string {
	s = strings.ReplaceAll(s, "@", "-at-")
	s = strings.ReplaceAll(s, "&", "-and-")
	return s
}

// Sanitize converts a filename (or text) to a URL-friendly format:
//   - URL decode
//   - Accent normalization
//   - Lowercase
//   - Special character replacements (@, &)
//   - Separators (spaces, parens, brackets, etc.) → hyphens
//   - 3+ consecutive hyphens → 2 hyphens
//   - Non-[a-z0-9-] characters removed
//   - Leading/trailing hyphens removed
//   - Leading dots, leading underscores, and up to 2 trailing extensions preserved
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

	// 3. Extract up to 2 trailing extensions (only alphanumeric segments)
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
		return -1 // drop
	}, name)

	// 10. Trim leading/trailing hyphens
	name = strings.Trim(name, "-")

	// 11. Lowercase the extension
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
