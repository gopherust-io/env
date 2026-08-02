package tag_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/gopherust-io/env/internal/tag"
)

const sample = `
package sample

type Database struct {
	Host string ` + "`env:\"HOST\" required:\"true\"`" + `
	Port int    ` + "`env:\"PORT\" default:\"5432\"`" + `
}

type Config struct {
	Port int      ` + "`env:\"PORT\" default:\"8080\"`" + `
	DB   Database ` + "`prefix:\"DB_\"`" + `
	Tags []string ` + "`env:\"TAGS\" sep:\",\"`" + `
}
`

func TestCollectFields(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "sample.go", sample, 0)
	if err != nil {
		t.Fatal(err)
	}

	st, _, err := tag.FindStruct([]*ast.File{f}, "Config")
	if err != nil {
		t.Fatal(err)
	}

	fields, err := tag.CollectFields(&tag.Context{Local: []*ast.File{f}}, st, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(fields) != 3 {
		t.Fatalf("expected 3 top-level fields, got %d", len(fields))
	}

	db := fields[1]
	if !db.IsNested || len(db.Children) != 2 {
		t.Fatalf("expected nested DB with 2 children, got %+v", db)
	}
	if db.Children[0].EnvKey != "HOST" || db.Prefix != "DB_" {
		t.Fatalf("unexpected nested child: %+v", db.Children[0])
	}
}

func TestCollectFieldsRejectsCyclicNestedTypes(t *testing.T) {
	const cyclic = `
package sample

type A struct {
	B B
}

type B struct {
	A A
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "cyclic.go", cyclic, 0)
	if err != nil {
		t.Fatal(err)
	}

	st, _, err := tag.FindStruct([]*ast.File{f}, "A")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tag.CollectFields(&tag.Context{Local: []*ast.File{f}}, st, "")
	if err == nil {
		t.Fatal("expected cyclic nested type error")
	}
	if !strings.Contains(err.Error(), "cyclic nested type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectFieldsRejectsEmbedded(t *testing.T) {
	const src = `
package sample

type Inner struct {
	Host string ` + "`env:\"HOST\"`" + `
}

type Config struct {
	Inner
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "embed.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	st, _, err := tag.FindStruct([]*ast.File{f}, "Config")
	if err != nil {
		t.Fatal(err)
	}
	_, err = tag.CollectFields(&tag.Context{Local: []*ast.File{f}}, st, "")
	if err == nil {
		t.Fatal("expected embedded field error")
	}
	if !strings.Contains(err.Error(), "embedded field not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchemaFieldsOmitsSensitiveDefault(t *testing.T) {
	schema := tag.SchemaFields([]tag.Field{
		{Name: "Password", EnvKey: "PASSWORD", Default: "s3cret", Sensitive: true, FieldPath: "Password"},
		{Name: "Port", EnvKey: "PORT", Default: "8080", FieldPath: "Port"},
	})
	if len(schema) != 2 {
		t.Fatalf("len=%d", len(schema))
	}
	if schema[0].Default != "" || !schema[0].HasDefault || !schema[0].Sensitive {
		t.Fatalf("sensitive schema entry: %+v", schema[0])
	}
	if schema[1].Default != "8080" || !schema[1].HasDefault {
		t.Fatalf("port schema entry: %+v", schema[1])
	}
}
