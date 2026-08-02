package env_test

import (
	"strings"
	"testing"

	"github.com/gopherust-io/env"
)

func TestLevenshteinClosestKey(t *testing.T) {
	keys := []string{"DB_HOST", "DB_PORT", "PORT"}
	got, dist := env.ClosestKey("DB_HST", keys)
	if got != "DB_HOST" || dist != 1 {
		t.Fatalf("ClosestKey(DB_HST)=%q,%d want DB_HOST,1", got, dist)
	}
}

func TestAuditTypoAndSilentDefault(t *testing.T) {
	schema := []env.SchemaField{
		{Key: "DB_HOST", FieldPath: "DB.Host", Required: true, Prefix: "DB_"},
		{Key: "HOST", FieldPath: "Host", HasDefault: true, Default: "localhost"},
		{Key: "API_TOKEN", FieldPath: "APIToken"},
	}
	snap := env.FromMap(map[string]string{
		"DB_HST":    "db.example",
		"API_TOKEN": "tok_example_abcdefghijklmnopqrstuvwxyz012345",
	})
	rep := env.Audit(snap, schema, env.AuditOptions{Mode: env.AuditModeProd})
	kinds := map[env.FindingKind]int{}
	for _, f := range rep.Findings {
		kinds[f.Kind]++
		switch f.Kind {
		case env.FindingTypo:
			if f.Suggest != "DB_HOST" || f.Severity != env.SeverityError {
				t.Fatalf("typo finding: %+v", f)
			}
		case env.FindingSilentDefault:
			if f.Severity != env.SeverityError {
				t.Fatalf("prod silent default should be error: %+v", f)
			}
		case env.FindingMissingRequired:
			if f.Key != "DB_HOST" {
				t.Fatalf("missing: %+v", f)
			}
		case env.FindingUnmarkedSecret:
			if f.Key != "API_TOKEN" {
				t.Fatalf("secret: %+v", f)
			}
		}
	}
	if kinds[env.FindingTypo] != 1 || kinds[env.FindingSilentDefault] != 1 ||
		kinds[env.FindingMissingRequired] != 1 || kinds[env.FindingUnmarkedSecret] != 1 {
		t.Fatalf("unexpected kinds: %v findings=%v", kinds, rep.Findings)
	}
	if err := rep.Err(); err == nil {
		t.Fatal("expected Err()")
	}
}

func TestAuditOrphanUnderPrefix(t *testing.T) {
	schema := []env.SchemaField{
		{Key: "DB_HOST", FieldPath: "DB.Host", Prefix: "DB_"},
	}
	snap := env.FromMap(map[string]string{
		"DB_POOL_SIZE": "10",
		"PATH":         "/bin",
	})
	rep := env.Audit(snap, schema, env.AuditOptions{})
	if len(rep.Findings) != 1 || rep.Findings[0].Kind != env.FindingOrphan {
		t.Fatalf("want one orphan, got %+v", rep.Findings)
	}
	if rep.Findings[0].Severity != env.SeverityWarning {
		t.Fatalf("orphan default severity warn, got %v", rep.Findings[0].Severity)
	}
	rep = env.Audit(snap, schema, env.AuditOptions{StrictUnknown: true})
	if rep.Findings[0].Severity != env.SeverityError {
		t.Fatal("strict orphan should be error")
	}
}

func TestAuditIgnoresUnrelatedProcessEnv(t *testing.T) {
	schema := []env.SchemaField{
		{Key: "APP_PORT", FieldPath: "Port", HasDefault: true, Default: "8080"},
	}
	snap := env.FromMap(map[string]string{
		"HOME": "/Users/me",
		"PATH": "/usr/bin",
	})
	rep := env.Audit(snap, schema, env.AuditOptions{Mode: env.AuditModeDev})
	for _, f := range rep.Findings {
		if f.Kind == env.FindingTypo {
			t.Fatalf("unexpected typo for process env: %+v", f)
		}
	}
	if len(rep.Findings) != 1 || rep.Findings[0].Kind != env.FindingSilentDefault {
		t.Fatalf("want only silent default, got %+v", rep.Findings)
	}
}

func TestLooksLikeSecret(t *testing.T) {
	if !env.LooksLikeSecret("sk_example_abc") {
		t.Fatal("prefix sk_")
	}
	if !env.LooksLikeSecret("abcdefghij0123456789KLMN") {
		t.Fatal("long mixed token")
	}
	if env.LooksLikeSecret("localhost") {
		t.Fatal("short non-secret")
	}
	if env.LooksLikeSecret("this is a sentence with spaces 12345") {
		t.Fatal("spaces should not look secret")
	}
}

func TestFindingKindString(t *testing.T) {
	if env.FindingTypo.String() != "typo" {
		t.Fatal(env.FindingTypo.String())
	}
	if !strings.Contains(env.SeverityError.String(), "error") {
		t.Fatal(env.SeverityError.String())
	}
}
