package env

import (
	"os"
	"sync/atomic"
)

// EnvSnapshot holds an indexed view of environment variables.
type EnvSnapshot struct {
	vars map[string]string
}

var defaultSnap atomic.Pointer[EnvSnapshot]

// Snapshot returns a cached process environment index.
// The index is built once from os.Environ() until ResetSnapshot is called.
func Snapshot() *EnvSnapshot {
	if s := defaultSnap.Load(); s != nil {
		return s
	}
	s := FromEnviron(os.Environ())
	if defaultSnap.CompareAndSwap(nil, s) {
		return s
	}
	return defaultSnap.Load()
}

// ResetSnapshot rebuilds the cached snapshot.
func ResetSnapshot() {
	defaultSnap.Store(FromEnviron(os.Environ()))
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
