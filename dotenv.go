package env

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
)

// MaxDotEnvBytes is the maximum .env file size accepted by ParseDotEnvFile.
const MaxDotEnvBytes = 1 << 20 // 1 MiB

// LoadDotEnv loads variables from path into the process environment.
// Existing variables are not overwritten. The cached snapshot is refreshed.
// path is treated as a trusted filesystem path.
func LoadDotEnv(path string) error {
	vars, err := ParseDotEnvFile(path)
	if err != nil {
		return err
	}
	var setErrs []error
	for key, val := range vars {
		if _, ok := os.LookupEnv(key); !ok {
			if err := os.Setenv(key, val); err != nil {
				setErrs = append(setErrs, fmt.Errorf("%s: %w", key, err))
			}
		}
	}
	ResetSnapshot()
	return errors.Join(setErrs...)
}

// ParseDotEnvFile reads KEY=VALUE pairs from a dotenv file.
func ParseDotEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(f, int64(MaxDotEnvBytes)+1))
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(data) > MaxDotEnvBytes {
		return nil, fmt.Errorf("dotenv: file %q exceeds %d bytes", path, MaxDotEnvBytes)
	}
	return ParseDotEnv(data)
}

// ParseDotEnv parses dotenv content. Supports # comments, export prefix, and quoted values.
func ParseDotEnv(data []byte) (map[string]string, error) {
	if len(data) > MaxDotEnvBytes {
		return nil, fmt.Errorf("dotenv: content exceeds %d bytes", MaxDotEnvBytes)
	}
	out := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), MaxDotEnvBytes)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if IsEmpty(line) || strings.HasPrefix(line, "#") {
			continue
		}
		line = stripUnquotedComment(line)
		if IsEmpty(line) {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		key = stripExportPrefix(key)
		val = strings.TrimSpace(val)
		if IsEmpty(key) {
			continue
		}
		val = unquoteDotEnvValue(val)
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func stripExportPrefix(key string) string {
	if !strings.HasPrefix(key, "export") {
		return key
	}
	rest := key[len("export"):]
	if rest == "" || !isASCIISpace(rest[0]) {
		return key
	}
	return strings.TrimLeft(rest, " \t")
}

// stripUnquotedComment removes a # comment that appears outside of quotes.
// Unquoted # is only treated as a comment when preceded by whitespace (so URL
// fragments like http://x.com#frag are preserved).
func stripUnquotedComment(line string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch c {
		case '\\':
			if inDouble && i+1 < len(line) {
				i++
			}
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble && (i == 0 || isASCIISpace(line[i-1])) {
				return strings.TrimSpace(line[:i])
			}
		}
	}
	return line
}

func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// unquoteDotEnvValue removes one matching quote pair from a dotenv value.
func unquoteDotEnvValue(val string) string {
	if len(val) < 2 {
		return val
	}
	q := val[0]
	if (q == '"' || q == '\'') && val[len(val)-1] == q {
		return val[1 : len(val)-1]
	}
	return val
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
