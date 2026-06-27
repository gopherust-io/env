package env_test

import (
	"testing"

	"github.com/gopherust-io/env"
)

func TestSnapshotLookup(t *testing.T) {
	snap := env.FromMap(map[string]string{
		"PORT":  "8080",
		"EMPTY": "",
	})

	v, ok := snap.Lookup("PORT")
	if !ok || v != "8080" {
		t.Fatalf("Lookup(PORT) = %q, %v", v, ok)
	}

	_, ok = snap.Lookup("MISSING")
	if ok {
		t.Fatal("expected missing key")
	}

	if snap.Len() != 2 {
		t.Fatalf("Len() = %d", snap.Len())
	}
}

func FuzzSnapshotLookup(f *testing.F) {
	f.Add("KEY", "VALUE")
	f.Fuzz(func(t *testing.T, key, value string) {
		snap := env.FromMap(map[string]string{key: value})
		got, ok := snap.Lookup(key)
		if !ok || got != value {
			t.Fatalf("Lookup(%q) = %q, %v", key, got, ok)
		}
	})
}
