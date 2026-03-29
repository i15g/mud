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
		{"RELEASE-NOTES", false}, // hyphen in stem — not protected
		{"FOO123", false},        // digits — not ^[A-Z]+$
		{"Foo", false},           // mixed case
		{"foo", false},           // lowercase

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
