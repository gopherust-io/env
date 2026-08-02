package env

import (
	"fmt"
	"strings"
	"unicode"
)

// FindingKind classifies an audit finding.
type FindingKind int

const (
	FindingTypo FindingKind = iota
	FindingOrphan
	FindingSilentDefault
	FindingMissingRequired
	FindingUnmarkedSecret
)

func (k FindingKind) String() string {
	switch k {
	case FindingTypo:
		return "typo"
	case FindingOrphan:
		return "orphan"
	case FindingSilentDefault:
		return "silent_default"
	case FindingMissingRequired:
		return "missing_required"
	case FindingUnmarkedSecret:
		return "unmarked_secret"
	default:
		return "unknown"
	}
}

// FindingSeverity is error or warning.
type FindingSeverity int

const (
	SeverityWarning FindingSeverity = iota
	SeverityError
)

func (s FindingSeverity) String() string {
	if s == SeverityError {
		return "error"
	}
	return "warn"
}

// AuditMode controls which warnings become errors.
type AuditMode int

const (
	// AuditModeDev keeps silent defaults and unmarked secrets as warnings.
	AuditModeDev AuditMode = iota
	// AuditModeProd elevates silent defaults and unmarked secrets to errors.
	AuditModeProd
)

// SchemaField describes one env key known to a generated (or CLI) schema.
type SchemaField struct {
	Key        string
	FieldPath  string
	Default    string
	Prefix     string
	Required   bool
	HasDefault bool
	Sensitive  bool
	_          [5]byte
}

// Finding is one contract violation or hygiene warning.
type Finding struct {
	Key      string
	Field    string
	Message  string
	Suggest  string
	Kind     FindingKind
	Severity FindingSeverity
}

// AuditReport collects findings from one audit pass.
type AuditReport struct {
	Findings []Finding
}

// Err returns a non-nil error when any finding has SeverityError.
func (r AuditReport) Err() error {
	var errs []string
	for _, f := range r.Findings {
		if f.Severity != SeverityError {
			continue
		}
		errs = append(errs, f.Message)
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("env audit: %s", strings.Join(errs, "; "))
}

// HasErrors reports whether any finding is SeverityError.
func (r AuditReport) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Severity == SeverityError {
			return true
		}
	}
	return false
}

// AuditOptions tunes contract checks.
type AuditOptions struct {
	Mode          AuditMode
	StrictUnknown bool // orphans are errors (default: warnings)
	AllUnknown    bool // report unknown keys that are neither orphans nor typos
	_             [6]byte
}

// Audit checks snap against schema: typos, orphans, silent defaults, required gaps, unmarked secrets.
func Audit(snap *EnvSnapshot, schema []SchemaField, opts AuditOptions) AuditReport {
	if snap == nil {
		snap = &EnvSnapshot{vars: map[string]string{}}
	}

	byKey := make(map[string]SchemaField, len(schema))
	prefixes := make(map[string]struct{})
	schemaKeys := make([]string, 0, len(schema))
	for _, f := range schema {
		byKey[f.Key] = f
		schemaKeys = append(schemaKeys, f.Key)
		if !IsEmpty(f.Prefix) {
			prefixes[f.Prefix] = struct{}{}
		}
	}

	var findings []Finding
	warnOrErr := SeverityWarning
	if opts.Mode == AuditModeProd {
		warnOrErr = SeverityError
	}

	for _, f := range schema {
		raw, ok := snap.Lookup(f.Key)
		present := ok && !IsEmpty(raw)
		if !present {
			if f.Required {
				findings = append(findings, Finding{
					Kind:     FindingMissingRequired,
					Severity: SeverityError,
					Key:      f.Key,
					Field:    f.FieldPath,
					Message:  fmt.Sprintf("%s (%s): required but unset", f.FieldPath, f.Key),
				})
			} else if f.HasDefault {
				defShow := f.Default
				if f.Sensitive {
					defShow = SensitiveMask
				}
				findings = append(findings, Finding{
					Kind:     FindingSilentDefault,
					Severity: warnOrErr,
					Key:      f.Key,
					Field:    f.FieldPath,
					Message:  fmt.Sprintf("%s (%s): unset; using default %q", f.FieldPath, f.Key, defShow),
				})
			}
			continue
		}
		if !f.Sensitive && LooksLikeSecret(raw) {
			findings = append(findings, Finding{
				Kind:     FindingUnmarkedSecret,
				Severity: warnOrErr,
				Key:      f.Key,
				Field:    f.FieldPath,
				Message:  fmt.Sprintf("%s (%s): value looks secret but lacks sensitive:\"true\"", f.FieldPath, f.Key),
			})
		}
	}

	snap.Range(func(key, _ string) bool {
		if _, known := byKey[key]; known {
			return true
		}
		if suggest, dist := ClosestKey(key, schemaKeys); dist >= 0 && dist <= maxTypoDistance(key, suggest) {
			findings = append(findings, Finding{
				Kind:     FindingTypo,
				Severity: SeverityError,
				Key:      key,
				Suggest:  suggest,
				Message:  fmt.Sprintf("%s is not a known key; did you mean %s?", key, suggest),
			})
			return true
		}
		if prefix, ok := matchingPrefix(key, prefixes); ok {
			sev := SeverityWarning
			if opts.StrictUnknown {
				sev = SeverityError
			}
			findings = append(findings, Finding{
				Kind:     FindingOrphan,
				Severity: sev,
				Key:      key,
				Message:  fmt.Sprintf("%s is not in schema but matches prefix %q", key, prefix),
			})
			return true
		}
		if opts.AllUnknown {
			sev := SeverityWarning
			if opts.StrictUnknown {
				sev = SeverityError
			}
			findings = append(findings, Finding{
				Kind:     FindingOrphan,
				Severity: sev,
				Key:      key,
				Message:  fmt.Sprintf("%s is not a known key", key),
			})
		}
		return true
	})

	return AuditReport{Findings: findings}
}

func matchingPrefix(key string, prefixes map[string]struct{}) (string, bool) {
	best := ""
	for p := range prefixes {
		if strings.HasPrefix(key, p) && len(p) > len(best) {
			best = p
		}
	}
	if IsEmpty(best) {
		return "", false
	}
	return best, true
}

func maxTypoDistance(a, b string) int {
	if min(len(a), len(b)) < 6 {
		return 1
	}
	return 2
}

// ClosestKey returns the schema key with minimal Levenshtein distance to key.
// Candidates whose length differs by more than the typo budget are skipped.
// Distance is -1 when no candidate remains.
func ClosestKey(key string, keys []string) (string, int) {
	best := ""
	bestDist := int(^uint(0) >> 1)
	for _, k := range keys {
		maxD := maxTypoDistance(key, k)
		if abs(len(key)-len(k)) > maxD {
			continue
		}
		d := levenshtein(key, k)
		if d < bestDist || (d == bestDist && (best == "" || k < best)) {
			bestDist = d
			best = k
		}
	}
	if best == "" {
		return "", -1
	}
	return best, bestDist
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if IsEmpty(a) {
		return len(b)
	}
	if IsEmpty(b) {
		return len(a)
	}
	// Ensure b is shorter row for slightly less memory.
	if len(a) < len(b) {
		a, b = b, a
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

// LooksLikeSecret reports whether value resembles an API token or secret.
func LooksLikeSecret(value string) bool {
	if IsEmpty(value) {
		return false
	}
	lower := strings.ToLower(value)
	for _, p := range secretPrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	if len(value) < 24 {
		return false
	}
	return highEntropyToken(value)
}

var secretPrefixes = []string{
	"sk_",
	"pk_",
	"rk_",
	"ghp_",
	"gho_",
	"ghu_",
	"ghs_",
	"ghr_",
	"xoxb-",
	"xoxp-",
	"xoxa-",
	"xoxs-",
	"akia",
}

func highEntropyToken(value string) bool {
	var letters, digits, other int
	for _, r := range value {
		switch {
		case unicode.IsLetter(r):
			letters++
		case unicode.IsDigit(r):
			digits++
		case r == '-' || r == '_' || r == '.' || r == '/' || r == '+' || r == '=':
			other++
		default:
			if unicode.IsSpace(r) {
				return false
			}
			other++
		}
	}
	n := len(value)
	// Token-like: mostly alnum / URL-safe, mixed classes, no spaces.
	if letters == 0 || digits == 0 {
		return false
	}
	if letters+digits+other != n {
		return false
	}
	return true
}
