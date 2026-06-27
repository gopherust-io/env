package codegen_test

import (
	"strings"
	"testing"

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
	out := string(src)
	for _, want := range []string{
		"func LoadConfig() (Config, error)",
		"func MustLoadConfig() Config",
		"env.ParseInt",
		"env.ParseBool",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}
