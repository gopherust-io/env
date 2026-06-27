package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherust-io/env/internal/codegen"
	"github.com/gopherust-io/env/internal/tag"
)

func main() {
	dir := flag.String("dir", ".", "package directory")
	typeName := flag.String("type", "", "struct type name")
	output := flag.String("output", "", "output file (default: <snake>_env_gen.go)")
	pkgName := flag.String("package", "", "package name override")
	module := flag.String("module", "github.com/gopherust-io/env", "env module import path")
	flag.Parse()

	if *typeName == "" {
		fmt.Fprintln(os.Stderr, "envgen: -type is required")
		os.Exit(2)
	}

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fatal(err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, absDir, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() &&
			!strings.HasSuffix(name, "_test.go") &&
			!strings.HasSuffix(name, "_env_gen.go") &&
			!strings.HasPrefix(name, "generate_")
	}, parser.ParseComments)
	if err != nil {
		fatal(err)
	}

	if len(pkgs) != 1 {
		fatal(fmt.Errorf("expected exactly one package in %s, found %d", absDir, len(pkgs)))
	}

	var pkg *ast.Package
	var pkgPath string
	for name, p := range pkgs {
		if strings.HasPrefix(name, "_") {
			continue
		}
		pkg = p
		pkgPath = name
		break
	}
	if pkg == nil {
		fatal(fmt.Errorf("no non-test package found in %s", absDir))
	}

	files := make([]*ast.File, 0, len(pkg.Files))
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	st, filePkg, err := tag.FindStruct(files, *typeName)
	if err != nil {
		fatal(err)
	}

	fields, err := tag.CollectFields(files, st, "")
	if err != nil {
		fatal(err)
	}

	outPkg := filePkg
	if *pkgName != "" {
		outPkg = *pkgName
	}

	outFile := *output
	if outFile == "" {
		outFile = toSnake(*typeName) + "_env_gen.go"
	}

	src, err := codegen.Generate(codegen.Options{
		Package:  outPkg,
		TypeName: *typeName,
		Fields:   fields,
		Module:   *module,
	})
	if err != nil {
		fatal(err)
	}

	outPath := filepath.Join(absDir, outFile)
	if err := os.WriteFile(outPath, src, 0o644); err != nil {
		fatal(err)
	}

	fmt.Printf("envgen: wrote %s for type %s in package %s\n", outPath, *typeName, pkgPath)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "envgen:", err)
	os.Exit(1)
}

func toSnake(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
