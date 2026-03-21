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
		{"My File (2024).txt", "my-file-2024.txt"},
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

		// Non-alphanumeric removal
		{"foo(bar)", "foobar"},
		{"café", "caf"},   // é removed
		{"naïve", "nave"}, // ï removed

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
