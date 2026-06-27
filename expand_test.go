package env_test

import (
	"os"
	"testing"
	"time"

	"github.com/gopherust-io/env"
)

func TestExpand(t *testing.T) {
	snap := env.FromMap(map[string]string{
		"HOST": "localhost",
		"PORT": "8080",
	})

	got := env.Expand("http://${HOST}:${PORT}/api", snap)
	if got != "http://localhost:8080/api" {
		t.Fatalf("Expand = %q", got)
	}

	got = env.Expand("$HOST:$PORT", snap)
	if got != "localhost:8080" {
		t.Fatalf("Expand = %q", got)
	}
}

func TestParseTime(t *testing.T) {
	got, err := env.ParseTime("2026-06-27T12:00:00Z", time.RFC3339)
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 {
		t.Fatalf("year = %d", got.Year())
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"
	if err := os.WriteFile(path, []byte("FOO=bar\n# comment\nBAZ=qux\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	env.ResetSnapshot()
	t.Setenv("FOO", "override")

	if err := env.LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}

	if v := lookupEnv("FOO"); v != "override" {
		t.Fatalf("FOO = %q", v)
	}
	if v := lookupEnv("BAZ"); v != "qux" {
		t.Fatalf("BAZ = %q", v)
	}
}

func TestSnapshotWithDotEnv(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"
	if err := os.WriteFile(path, []byte("FOO=file\nBAR=fromfile\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("BAR", "fromenv")

	snap, err := env.SnapshotWithDotEnv(path)
	if err != nil {
		t.Fatal(err)
	}

	v, _ := snap.Lookup("FOO")
	if v != "file" {
		t.Fatalf("FOO = %q", v)
	}
	v, _ = snap.Lookup("BAR")
	if v != "fromenv" {
		t.Fatalf("BAR = %q", v)
	}
}

func lookupEnv(key string) string {
	v, _ := env.Snapshot().Lookup(key)
	return v
}
