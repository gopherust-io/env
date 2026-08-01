package bench

import (
	"github.com/gopherust-io/env"
	"os"
	"strconv"
	"time"
)

func lookupString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && !env.IsEmpty(v) {
		return v
	}
	return def
}

func lookupInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && !env.IsEmpty(v) {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func lookupBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && !env.IsEmpty(v) {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func lookupDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && !env.IsEmpty(v) {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func lookupStringEmpty(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return ""
}
