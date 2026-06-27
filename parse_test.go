package env_test

import (
	"testing"
	"time"

	"github.com/gopherust-io/env"
)

func TestParseBool(t *testing.T) {
	tests := map[string]bool{
		"true":  true,
		"false": false,
		"1":     true,
		"0":     false,
		"yes":   true,
		"no":    false,
	}
	for in, want := range tests {
		got, err := env.ParseBool(in)
		if err != nil || got != want {
			t.Fatalf("ParseBool(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
}

func TestParseIntSlice(t *testing.T) {
	got, err := env.ParseIntSlice("1,2,3", ",")
	if err != nil {
		t.Fatal(err)
	}
	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestParseStringMap(t *testing.T) {
	got, err := env.ParseStringMap("a:1,b:2", ",", ":")
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != "1" || got["b"] != "2" {
		t.Fatalf("unexpected map: %v", got)
	}
}

func TestParseDuration(t *testing.T) {
	got, err := env.ParseDuration("10s")
	if err != nil {
		t.Fatal(err)
	}
	if got != 10*time.Second {
		t.Fatalf("got %v", got)
	}
}

func FuzzParseInt(f *testing.F) {
	f.Add("42")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = env.ParseInt(s)
	})
}

func FuzzParseBool(f *testing.F) {
	f.Add("true")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = env.ParseBool(s)
	})
}
