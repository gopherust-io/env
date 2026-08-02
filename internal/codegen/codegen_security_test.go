package codegen_test

import (
	"strings"
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/internal/codegen"
	"github.com/gopherust-io/env/internal/tag"
)

func TestGenerateDeepNested(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{
			{
				Name:     "M",
				IsNested: true,
				Prefix:   "M_",
				Children: []tag.Field{
					{
						Name:     "In",
						IsNested: true,
						Prefix:   "M_IN_",
						Children: []tag.Field{
							{Name: "Host", GoType: "string", EnvKey: "HOST", Prefix: "M_IN_", FieldPath: "M.In.Host"},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	for _, want := range []string{
		`key := "M_IN_HOST"`,
		"cfg.M.In.Host = raw",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "cfg.M.In = ") && !strings.Contains(out, "cfg.M.In.Host") {
		t.Fatalf("nested child treated as scalar:\n%s", out)
	}
}

func TestGeneratePointerNestAndScalar(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{
			{
				Name:      "DB",
				GoType:    "*Database",
				IsNested:  true,
				IsPointer: true,
				Prefix:    "DB_",
				Children: []tag.Field{
					{Name: "Host", GoType: "string", EnvKey: "HOST", Prefix: "DB_", FieldPath: "DB.Host"},
				},
			},
			{Name: "Port", GoType: "*int", EnvKey: "PORT", IsPointer: true, FieldPath: "Port"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	for _, want := range []string{
		"if cfg.DB == nil {",
		"cfg.DB = &Database{}",
		"cfg.DB.Host = raw",
		"cfg.Port = &v",
		"env.ParseInt(raw)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "cfg.Port = v\n") {
		t.Fatalf("assigned int to *int:\n%s", out)
	}
}

func TestGenerateSensitiveAppend(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{
			{Name: "Password", GoType: "string", EnvKey: "PASSWORD", Sensitive: true, FieldPath: "Password"},
			{Name: "Token", GoType: "*string", EnvKey: "TOKEN", Sensitive: true, IsPointer: true, FieldPath: "Token"},
			{Name: "Secret", GoType: "string", EnvKey: "SECRET", Sensitive: true, Unmarshal: true, FieldPath: "Secret"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	if !strings.Contains(out, "AppendParseSensitive") {
		t.Fatalf("expected AppendParseSensitive:\n%s", out)
	}
	if !strings.Contains(out, "out.Password = env.SensitiveMask") {
		t.Fatalf("expected string mask:\n%s", out)
	}
	if !strings.Contains(out, "out.Token = &v") {
		t.Fatalf("expected *string mask:\n%s", out)
	}
}

func TestGenerateRejectsSensitiveNonString(t *testing.T) {
	_, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields:   []tag.Field{{Name: "N", GoType: "int", EnvKey: "N", Sensitive: true, FieldPath: "N"}},
	})
	if err == nil {
		t.Fatal("expected error for sensitive int")
	}
}

func TestGenerateRejectsSensitiveDefault(t *testing.T) {
	_, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{{
			Name: "Password", GoType: "string", EnvKey: "PASSWORD",
			Sensitive: true, Default: "s3cret", FieldPath: "Password",
		}},
	})
	if err == nil {
		t.Fatal("expected error for sensitive default")
	}
}

func TestGenerateMaskedClonesPointerNest(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{{
			Name: "DB", GoType: "*Database", IsNested: true, IsPointer: true, Prefix: "DB_",
			Children: []tag.Field{
				{Name: "Password", GoType: "string", EnvKey: "PASSWORD", Prefix: "DB_", Sensitive: true, FieldPath: "DB.Password"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	for _, want := range []string{
		"tmp := *out.DB",
		"out.DB = &tmp",
		"out.DB.Password = env.SensitiveMask",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestGenerateUnmarshalNilPointer(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{{
			Name: "M", GoType: "*Mode", EnvKey: "MODE", IsPointer: true, Unmarshal: true, FieldPath: "M",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	for _, want := range []string{
		"if cfg.M == nil {",
		"cfg.M = new(Mode)",
		"cfg.M.UnmarshalEnv(key, raw)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestGenerateExpandDefaultNonString(t *testing.T) {
	src, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{{
			Name: "Port", GoType: "int", EnvKey: "PORT", Default: "${PORT_DEFAULT}", Expand: true, FieldPath: "Port",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := env.BytesToString(src)
	if !strings.Contains(out, `defaultRaw := env.Expand("${PORT_DEFAULT}", snap)`) {
		t.Fatalf("missing expand default:\n%s", out)
	}
	if !strings.Contains(out, "env.ParseInt(defaultRaw)") {
		t.Fatalf("missing ParseInt(defaultRaw):\n%s", out)
	}
}

func TestGenerateRejectsUnsupportedSlice(t *testing.T) {
	_, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{{
			Name: "Xs", GoType: "[]float64", EnvKey: "XS", IsSlice: true, SliceElem: "float64", FieldPath: "Xs",
		}},
	})
	if err == nil {
		t.Fatal("expected unsupported slice error")
	}
}

func TestGenerateRejectsUnmarshalDefault(t *testing.T) {
	_, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields: []tag.Field{{
			Name: "M", GoType: "Mode", EnvKey: "MODE", Unmarshal: true, Default: "x", FieldPath: "M",
		}},
	})
	if err == nil {
		t.Fatal("expected unmarshal+default error")
	}
}

func TestGenerateRejectsBadModule(t *testing.T) {
	_, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Module:   `github.com/x/y"` + "\n\t" + `"os`,
		Fields:   []tag.Field{{Name: "Port", GoType: "int", EnvKey: "PORT", FieldPath: "Port"}},
	})
	if err == nil {
		t.Fatal("expected invalid module error")
	}
}

func TestGenerateRejectsUnsupportedType(t *testing.T) {
	_, err := codegen.Generate(codegen.Options{
		Package:  "sample",
		TypeName: "Config",
		Fields:   []tag.Field{{Name: "C", GoType: "complex128", EnvKey: "C", FieldPath: "C"}},
	})
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
}
