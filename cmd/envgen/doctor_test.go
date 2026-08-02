package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gopherust-io/env"
)

func TestRunDoctorFindsTypo(t *testing.T) {
	dir := t.TempDir()
	src := `package cfg

type Config struct {
	Host   string ` + "`env:\"HOST\" default:\"localhost\"`" + `
	DBHost string ` + "`env:\"DB_HOST\" required:\"true\"`" + `
}
`
	if err := os.WriteFile(filepath.Join(dir, "config.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/cfg\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("DB_HST=db.example\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := runDoctor([]string{
		"-dir", dir,
		"-type", "Config",
		"-env-file", envPath,
		"-mode", "prod",
		"-prefix", "DB_",
	})
	_ = w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	if code != 1 {
		t.Fatalf("exit code=%d want 1; out=%q", code, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "DB_HST") || !strings.Contains(out, "DB_HOST") {
		t.Fatalf("expected typo finding, got %q", out)
	}
}

func TestWriteDoctorReportJSON(t *testing.T) {
	var buf bytes.Buffer
	rep := env.AuditReport{Findings: []env.Finding{{
		Kind:     env.FindingTypo,
		Severity: env.SeverityError,
		Key:      "DB_HST",
		Suggest:  "DB_HOST",
		Message:  "DB_HST is not a known key; did you mean DB_HOST?",
	}}}
	if err := writeDoctorReport(&buf, rep, "json"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"kind": "typo"`) {
		t.Fatalf("json=%s", buf.String())
	}
}

func TestParseAuditMode(t *testing.T) {
	if _, err := parseAuditMode("nope"); err == nil {
		t.Fatal("expected error")
	}
	m, err := parseAuditMode("prod")
	if err != nil || m != env.AuditModeProd {
		t.Fatalf("got %v %v", m, err)
	}
}
