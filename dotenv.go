package env

import (
	"bufio"
	"bytes"
	"maps"
	"os"
	"strings"
)

// LoadDotEnv loads variables from path into the process environment.
// Existing variables are not overwritten. The cached snapshot is refreshed.
func LoadDotEnv(path string) error {
	vars, err := ParseDotEnvFile(path)
	if err != nil {
		return err
	}
	for key, val := range vars {
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, val)
		}
	}
	ResetSnapshot()
	return nil
}

// ParseDotEnvFile reads KEY=VALUE pairs from a dotenv file.
func ParseDotEnvFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseDotEnv(data)
}

// ParseDotEnv parses dotenv content. Supports # comments and quoted values.
func ParseDotEnv(data []byte) (map[string]string, error) {
	out := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		val = strings.Trim(val, `"'`)
		out[key] = val
	}
	return out, scanner.Err()
}

// SnapshotWithDotEnv builds a snapshot from dotenv files overlaid with os.Environ().
// Process environment values take precedence over file values.
func SnapshotWithDotEnv(paths ...string) (*EnvSnapshot, error) {
	vars := make(map[string]string)
	for _, path := range paths {
		fileVars, err := ParseDotEnvFile(path)
		if err != nil {
			return nil, err
		}
		maps.Copy(vars, fileVars)
	}
	for _, kv := range os.Environ() {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				vars[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return FromMap(vars), nil
}
