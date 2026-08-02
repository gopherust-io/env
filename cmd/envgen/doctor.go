package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/internal/tag"
)

func runDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", ".", "package directory")
	typeName := fs.String("type", "", "struct type name to audit")
	envFile := fs.String("env-file", "", "optional .env file to merge into the snapshot")
	mode := fs.String("mode", "dev", "audit mode: dev or prod")
	format := fs.String("format", "text", "output format: text or json")
	prefix := fs.String("prefix", "", "only consider env keys with this prefix for unknown-key checks")
	strictUnknown := fs.Bool("strict-unknown", false, "treat orphan/unknown keys as errors")
	allUnknown := fs.Bool("all-unknown", false, "report unknown keys that are neither orphans nor typos")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if env.IsEmpty(*typeName) {
		fmt.Fprintln(os.Stderr, "envgen doctor: -type is required")
		return 2
	}
	auditMode, err := parseAuditMode(*mode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 2
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(os.Stderr, "envgen doctor: invalid -format %q (want text or json)\n", *format)
		return 2
	}

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 1
	}
	ctx, _, err := loadPackageContext(absDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 1
	}
	st, _, err := tag.FindStruct(ctx.Local, *typeName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 1
	}
	fields, err := tag.CollectFields(ctx, st, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 1
	}
	schema := tag.SchemaFields(fields)

	snap, err := doctorSnapshot(*envFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 1
	}
	if !env.IsEmpty(*prefix) {
		snap = filterSnapshotPrefix(snap, *prefix)
		schema = filterSchemaPrefix(schema, *prefix)
	}

	report := env.Audit(snap, schema, env.AuditOptions{
		Mode:          auditMode,
		StrictUnknown: *strictUnknown,
		AllUnknown:    *allUnknown,
	})

	if err := writeDoctorReport(os.Stdout, report, *format); err != nil {
		fmt.Fprintln(os.Stderr, "envgen doctor:", err)
		return 1
	}
	if report.HasErrors() {
		return 1
	}
	return 0
}

func parseAuditMode(mode string) (env.AuditMode, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "dev":
		return env.AuditModeDev, nil
	case "prod":
		return env.AuditModeProd, nil
	default:
		return 0, fmt.Errorf("invalid -mode %q (want dev or prod)", mode)
	}
}

func doctorSnapshot(envFile string) (*env.EnvSnapshot, error) {
	if env.IsEmpty(envFile) {
		return env.Snapshot(), nil
	}
	return env.SnapshotWithDotEnv(envFile)
}

func filterSnapshotPrefix(snap *env.EnvSnapshot, prefix string) *env.EnvSnapshot {
	m := make(map[string]string)
	snap.Range(func(key, value string) bool {
		if strings.HasPrefix(key, prefix) {
			m[key] = value
		}
		return true
	})
	return env.FromMap(m)
}

func filterSchemaPrefix(schema []env.SchemaField, prefix string) []env.SchemaField {
	out := make([]env.SchemaField, 0, len(schema))
	for _, f := range schema {
		if strings.HasPrefix(f.Key, prefix) {
			out = append(out, f)
		}
	}
	return out
}

func writeDoctorReport(w io.Writer, report env.AuditReport, format string) error {
	if format == "json" {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		type findingJSON struct {
			Kind     string `json:"kind"`
			Severity string `json:"severity"`
			Key      string `json:"key,omitempty"`
			Field    string `json:"field,omitempty"`
			Message  string `json:"message"`
			Suggest  string `json:"suggest,omitempty"`
		}
		out := struct {
			Findings []findingJSON `json:"findings"`
			Ok       bool          `json:"ok"`
		}{
			Findings: make([]findingJSON, 0, len(report.Findings)),
			Ok:       !report.HasErrors(),
		}
		for _, f := range report.Findings {
			out.Findings = append(out.Findings, findingJSON{
				Kind:     f.Kind.String(),
				Severity: f.Severity.String(),
				Key:      f.Key,
				Field:    f.Field,
				Message:  f.Message,
				Suggest:  f.Suggest,
			})
		}
		return enc.Encode(out)
	}

	if len(report.Findings) == 0 {
		_, err := fmt.Fprintln(w, "ok: no findings")
		return err
	}
	for _, f := range report.Findings {
		if _, err := fmt.Fprintf(w, "%s: %s\n", f.Severity.String(), f.Message); err != nil {
			return err
		}
	}
	return nil
}
