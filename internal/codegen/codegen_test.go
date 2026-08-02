package codegen_test

import (
	"strings"
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/internal/codegen"
	"github.com/gopherust-io/env/internal/tag"
)

func TestGenerateFlatStruct(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{
			{Name: "Port", GoType: "int", EnvKey: "PORT", Default: "8080", FieldPath: "Port"},
			{Name: "Debug", GoType: "bool", EnvKey: "DEBUG", FieldPath: "Debug"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	for _, want := range []string{
		"func LoadConfig() (Config, error)",
		"func ReloadConfig(cfg *Config) error",
		"func LoadConfigFrom(snap *env.EnvSnapshot) (Config, error)",
		"func MustLoadConfig() Config",
		"func AuditConfig(snap *env.EnvSnapshot) env.AuditReport",
		"func AuditConfigFromEnviron() env.AuditReport",
		"configAuditSchema",
		`Key: "PORT"`,
		"loadConfig(&cfg, env.Snapshot())",
		"env.ParseInt",
		"env.ParseBool",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Count(out, "env.Reload()") != 1 {
		t.Fatalf("expected env.Reload only in ReloadConfig, got:\n%s", out)
	}
	if !strings.Contains(out, "var next Config") || !strings.Contains(out, "*cfg = next") {
		t.Fatalf("ReloadConfig must load into a fresh value:\n%s", out)
	}
	loadIdx := strings.Index(out, "func LoadConfig()")
	reloadIdx := strings.Index(out, "func ReloadConfig(")
	if loadIdx < 0 || reloadIdx < 0 || reloadIdx < loadIdx {
		t.Fatal("unexpected Load/Reload ordering")
	}
	loadBody := out[loadIdx:reloadIdx]
	if strings.Contains(loadBody, "env.Reload()") {
		t.Fatalf("LoadConfig must not call env.Reload():\n%s", loadBody)
	}
}
