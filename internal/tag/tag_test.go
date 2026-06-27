package tag_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/gopherust-io/env/internal/tag"
)

const sample = `
package sample

type Database struct {
	Host string ` + "`env:\"HOST\" required`" + `
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

	fields, err := tag.CollectFields([]*ast.File{f}, st, "")
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
