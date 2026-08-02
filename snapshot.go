package env

import (
	"maps"
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

const snapFlightKey = "env-snapshot"

// Snapshot returns a cached process environment index.
// The index is built once from os.Environ() until ResetSnapshot is called.
// Concurrent first callers share a single FromEnviron via singleflight.
func Snapshot() *EnvSnapshot {
	if s := defaultSnap.Load(); s != nil {
		return s
	}
	v, _, _ := snapFlight.Do(snapFlightKey, func() (any, error) {
		if s := defaultSnap.Load(); s != nil {
			return s, nil
		}
		s := FromEnviron(os.Environ())
		// A concurrent ResetSnapshot may have stored a fresher snapshot already.
		if !defaultSnap.CompareAndSwap(nil, s) {
			if cur := defaultSnap.Load(); cur != nil {
				return cur, nil
			}
		}
		return s, nil
	})
	if s, ok := v.(*EnvSnapshot); ok && s != nil {
		return s
	}
	if s := defaultSnap.Load(); s != nil {
		return s
	}
	return FromEnviron(os.Environ())
}

// ResetSnapshot rebuilds the cached snapshot from the current process environment.
// It always stores a fresh snapshot and does not share work with Snapshot's first-build
// flight, so a concurrent first Snapshot cannot overwrite a newer reload.
func ResetSnapshot() {
	defaultSnap.Store(FromEnviron(os.Environ()))
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

// FromMap copies vars into a new snapshot. Callers may mutate the input map afterward.
func FromMap(vars map[string]string) *EnvSnapshot {
	if vars == nil {
		return &EnvSnapshot{vars: map[string]string{}}
	}
	return &EnvSnapshot{vars: maps.Clone(vars)}
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

// Range calls fn for each key/value. If fn returns false, iteration stops.
func (s *EnvSnapshot) Range(fn func(key, value string) bool) {
	if s == nil || fn == nil {
		return
	}
	for k, v := range s.vars {
		if !fn(k, v) {
			return
		}
	}
}
