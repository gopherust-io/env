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
	files, sourcePkg, err := loadPackageFiles(fset, absDir)
	if err != nil {
		fatal(err)
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

	fmt.Printf("envgen: wrote %s for type %s in package %s\n", outPath, *typeName, sourcePkg)
}

func loadPackageFiles(fset *token.FileSet, dir string) ([]*ast.File, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}

	var files []*ast.File
	var pkgName string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") ||
			strings.HasSuffix(name, "_env_gen.go") ||
			strings.HasPrefix(name, "generate_") {
			continue
		}

		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", name, err)
		}
		if pkgName == "" {
			pkgName = file.Name.Name
		} else if file.Name.Name != pkgName {
			return nil, "", fmt.Errorf("multiple packages in %s: %s and %s", dir, pkgName, file.Name.Name)
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		return nil, "", fmt.Errorf("no Go source files in %s", dir)
	}

	return files, pkgName, nil
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
