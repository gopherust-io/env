package env_test

import (
	"testing"

	"github.com/gopherust-io/env"
)

func TestExpandLeavesMissingRefs(t *testing.T) {
	snap := env.FromMap(map[string]string{"HOST": "localhost"})
	got := env.Expand("x=${MISSING}y", snap)
	if got != "x=${MISSING}y" {
		t.Fatalf("got %q", got)
	}
	got = env.Expand("a=$MISSING/b", snap)
	if got != "a=$MISSING/b" {
		t.Fatalf("got %q", got)
	}
	got = env.Expand("${HOST}/ok", snap)
	if got != "localhost/ok" {
		t.Fatalf("got %q", got)
	}
}
