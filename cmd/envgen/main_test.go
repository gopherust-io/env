package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveOutputPathContained(t *testing.T) {
	dir := t.TempDir()
	got, err := resolveOutputPath(dir, "config_env_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "config_env_gen.go")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveOutputPathRejectsEscape(t *testing.T) {
	dir := t.TempDir()
	_, err := resolveOutputPath(dir, "../../../tmp/evil.go")
	if err == nil {
		t.Fatal("expected escape error")
	}
	if !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestResolveOutputPathRejectsAbsolute(t *testing.T) {
	dir := t.TempDir()
	_, err := resolveOutputPath(dir, filepath.Join(dir, "x.go"))
	if err == nil {
		t.Fatal("expected absolute path error")
	}
}
