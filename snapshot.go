package env

import (
	"os"
	"sync/atomic"

	"golang.org/x/sync/singleflight"
)

// EnvSnapshot holds an indexed view of environment variables.
type EnvSnapshot struct {
	vars map[string]string
}

var (
	defaultSnap atomic.Pointer[EnvSnapshot]
	snapFlight  singleflight.Group
)

// Snapshot returns a cached process environment index.
// The index is built once from os.Environ() until ResetSnapshot is called.
// Concurrent first callers share a single FromEnviron via singleflight.
func Snapshot() *EnvSnapshot {
	if s := defaultSnap.Load(); s != nil {
		return s
	}
	v, _, _ := snapFlight.Do("snapshot", func() (any, error) {
		if s := defaultSnap.Load(); s != nil {
			return s, nil
		}
		s := FromEnviron(os.Environ())
		defaultSnap.Store(s)
		return s, nil
	})
	return v.(*EnvSnapshot)
}

// ResetSnapshot rebuilds the cached snapshot.
// Concurrent Reload/ResetSnapshot calls coalesce to one environ rebuild.
func ResetSnapshot() {
	_, _, _ = snapFlight.Do("reload", func() (any, error) {
		defaultSnap.Store(FromEnviron(os.Environ()))
		return nil, nil
	})
}

// Reload refreshes the cached snapshot from os.Environ().
// Generated ReloadConfig calls this before re-parsing.
func Reload() {
	ResetSnapshot()
}

func FromEnviron(environ []string) *EnvSnapshot {
	m := make(map[string]string, len(environ))
	for _, kv := range environ {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				m[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return &EnvSnapshot{vars: m}
}

func FromMap(vars map[string]string) *EnvSnapshot {
	return &EnvSnapshot{vars: vars}
}

func (s *EnvSnapshot) Lookup(key string) (string, bool) {
	if s == nil {
		return "", false
	}
	v, ok := s.vars[key]
	return v, ok
}

func (s *EnvSnapshot) Len() int {
	if s == nil {
		return 0
	}
	return len(s.vars)
}
