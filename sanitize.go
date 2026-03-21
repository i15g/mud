package main

import "strings"

// Sanitize converts a filename (or text) to a URL-friendly format:
//   - Lowercase
//   - Spaces, newlines, underscores, commas, plus signs → hyphens
//   - 3+ consecutive hyphens → 2 hyphens
//   - Non-[a-z0-9.-] characters removed
//   - Leading/trailing hyphens removed
//   - Leading dots, leading underscores, and last file extension preserved
func Sanitize(name string) string {
	// 1. Strip leading dots
	var leadingDots string
	for strings.HasPrefix(name, ".") {
		leadingDots += "."
		name = name[1:]
	}

	// 2. Extract last extension (only from remaining name, after dots stripped)
	var ext string
	if i := strings.LastIndex(name, "."); i >= 0 {
		ext = name[i:] // includes the dot
		name = name[:i]
	}

	// 3. Strip leading underscores from stem
	var leadingUnderscores string
	for strings.HasPrefix(name, "_") {
		leadingUnderscores += "_"
		name = name[1:]
	}

	// 4. Lowercase
	name = strings.ToLower(name)

	// 5. Replace [ \n_,+] with hyphens (byte-level, all ASCII)
	name = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '_', ',', '+':
			return '-'
		}
		return r
	}, name)

	// 6. Collapse 3+ consecutive hyphens to 2
	name = collapseHyphens(name)

	// 7. Remove chars not in [a-z0-9.-]
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
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
