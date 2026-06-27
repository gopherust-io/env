package env

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Expand replaces ${VAR} and $VAR references using values from snap.
func Expand(s string, snap *EnvSnapshot) string {
	if s == "" || snap == nil {
		return s
	}
	if !strings.Contains(s, "$") {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] != '$' {
			b.WriteByte(s[i])
			continue
		}

		if i+1 < len(s) && s[i+1] == '{' {
			end := strings.IndexByte(s[i+2:], '}')
			if end < 0 {
				b.WriteByte(s[i])
				continue
			}
			key := s[i+2 : i+2+end]
			if v, ok := snap.Lookup(key); ok {
				b.WriteString(v)
			}
			i += 2 + end
			continue
		}

		j := i + 1
		for j < len(s) {
			r, size := utf8.DecodeRuneInString(s[j:])
			if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
				j += size
				continue
			}
			break
		}
		if j == i+1 {
			b.WriteByte(s[i])
			continue
		}
		key := s[i+1 : j]
		if v, ok := snap.Lookup(key); ok {
			b.WriteString(v)
		}
		i = j - 1
	}

	return b.String()
}
