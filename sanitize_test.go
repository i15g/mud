package main

import "testing"

func TestSanitize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Examples from help text
		{"FOO", "foo"},
		{"foo bar", "foo-bar"},
		{"foo_bar", "foo-bar"},
		{".bashrc", ".bashrc"},
		{".foo bar", ".foo-bar"},
		{"_private file", "_private-file"},
		{"My File (2024).txt", "my-file--2024.txt"},
		{"foo.bar.baz", "foo.bar.baz"},
		{"foo---bar", "foo--bar"},
		{"hello world 2024", "hello-world-2024"},

		// Lowercase
		{"UPPER.TXT", "upper.txt"},
		{"MixedCase.JPG", "mixedcase.jpg"},

		// Already clean
		{"already-clean", "already-clean"},
		{"foo-bar.txt", "foo-bar.txt"},

		// Underscores
		{"foo_bar_baz", "foo-bar-baz"},
		{"__double", "__double"},
		{"__private file", "__private-file"},

		// Commas and plus
		{"a,b", "a-b"},
		{"a+b", "a-b"},
		{"a,,b", "a--b"},
		{"a,,,b", "a--b"}, // 3 commas → 3 hyphens → collapsed to 2
		{"a+++b", "a--b"}, // 3 plus → 3 hyphens → collapsed to 2

		// Hyphen collapsing
		{"foo----bar", "foo--bar"},
		{"foo-----bar", "foo--bar"},
		{"a--b", "a--b"}, // 2 is fine, not collapsed

		// Leading/trailing hyphens removed
		{"-foo-", "foo"},
		{"--foo--", "foo"},

		// Non-alphanumeric removal (v1 behavior - accents stripped)
		// In v2, accents are normalized instead of removed
		{"café", "cafe"},   // é → e
		{"naïve", "naive"}, // ï → i

		// Newline
		{"a\nb", "a-b"},

		// Dotfiles
		{".bashrc", ".bashrc"},
		{".foo-bar", ".foo-bar"},
		{"..hidden", "..hidden"},
		{"...secret", "...secret"},

		// Dotfile with spaces
		{".foo bar baz", ".foo-bar-baz"},

		// Extension handling
		{"My File.TXT", "my-file.txt"},
		{"foo.bar.BAZ", "foo.bar.baz"},
		{"noext", "noext"},

		// Mixed: dotfile + underscore + extension
		{"._foo bar.TXT", "._foo-bar.txt"},

		// Empty after sanitize (edge case: input is just hyphens)
		{"---", ""},
		{"-", ""},

		// Multi-extension (v2)
		{"foo.tar.gz", "foo.tar.gz"},
		{"My.Config.File.txt", "my-config.file.txt"},
		{"hello world.foo bar baz.txt", "hello-world-foo-bar-baz.txt"},
		{"archive.tar.GZ", "archive.tar.gz"},
		{"_.txt", "_.txt"},

		// New separators (v2)
		{"foo(bar).txt", "foo-bar.txt"},
		{"foo[1].txt", "foo-1.txt"},
		{"a;b", "a-b"},
		{"a:b", "a-b"},
		{"a|b", "a-b"},
		{"a\u2014b", "a-b"}, // em dash
		{"a\u2013b", "a-b"}, // en dash
		{"a{b}", "a-b"},

		// URL decoding (v2 step 0)
		{"hello%20world.txt", "hello-world.txt"},
		{"foo%28bar%29.txt", "foo-bar.txt"},

		// Accent normalization (v2 step 1)
		{"\u00c9tude.txt", "etude.txt"},   // É → e
		{"stra\u00dfe", "strasse"},        // ß → ss
		{"pi\u00f1ata", "pinata"},         // ñ → n
		{"fa\u00e7ade", "facade"},         // ç → c
		{"\u00e0 la carte", "a-la-carte"}, // à → a

		// Special replacements (v2 step 6)
		{"foo@bar", "foo-at-bar"},
		{"a&b", "a-and-b"},
		{"@foo.txt", "at-foo.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
