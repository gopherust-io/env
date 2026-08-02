package env_test

import (
	"sync"
	"testing"

	"github.com/gopherust-io/env"
)

func TestFromMapClones(t *testing.T) {
	m := map[string]string{"K": "v"}
	snap := env.FromMap(m)
	m["K"] = "mutated"
	got, ok := snap.Lookup("K")
	if !ok || got != "v" {
		t.Fatalf("Lookup after mutate = %q, %v", got, ok)
	}
}

func TestResetSnapshotNotOverwrittenByFirstBuild(t *testing.T) {
	t.Setenv("SNAP_RACE_KEY", "from-env")
	env.ResetSnapshot()

	var wg sync.WaitGroup
	const n = 32
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			_ = env.Snapshot()
			env.ResetSnapshot()
			_ = env.Snapshot()
		}()
	}
	wg.Wait()

	t.Setenv("SNAP_RACE_KEY", "after-reset")
	env.ResetSnapshot()
	got, ok := env.Snapshot().Lookup("SNAP_RACE_KEY")
	if !ok || got != "after-reset" {
		t.Fatalf("Lookup = %q, %v", got, ok)
	}
}
