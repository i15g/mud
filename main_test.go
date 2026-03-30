package main

import (
	"bytes"
	"strings"
	"testing"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	old := stderr
	stderr = &buf
	defer func() { stderr = old }()
	fn()
	return buf.String()
}

func TestRun_HelpShowsVersion(t *testing.T) {
	out := captureStderr(t, func() {
		run([]string{"--help"}, strings.NewReader(""), false)
	})
	if !strings.Contains(out, version) {
		t.Errorf("help output missing version %q: %q", version, out)
	}
}

func TestRun_Version(t *testing.T) {
	out := captureStderr(t, func() {
		code := run([]string{"--version"}, strings.NewReader(""), false)
		if code != 0 {
			t.Fatalf("expected exit 0, got %d", code)
		}
	})
	if !strings.Contains(out, "mud") {
		t.Errorf("version output missing 'mud': %q", out)
	}
}
