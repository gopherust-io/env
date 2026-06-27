package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

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

	ctx, sourcePkg, err := loadPackageContext(absDir)
	if err != nil {
		fatal(err)
	}

	st, filePkg, err := tag.FindStruct(ctx.Local, *typeName)
	if err != nil {
		fatal(err)
	}

	fields, err := tag.CollectFields(ctx, st, "")
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

func loadPackageContext(dir string) (*tag.Context, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedSyntax | packages.NeedModule | packages.NeedDeps | packages.NeedFiles | packages.NeedImports,
		Dir:   absDir,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, "", err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, "", fmt.Errorf("package load failed in %s", absDir)
	}

	var main *packages.Package
	for _, p := range pkgs {
		if filepath.Clean(p.Dir) == filepath.Clean(absDir) && len(p.Syntax) > 0 {
			main = p
			break
		}
	}
	if main == nil {
		return nil, "", fmt.Errorf("no package found in %s", absDir)
	}

	deps := make(map[string][]*ast.File)
	for path, imp := range main.Imports {
		if imp == nil || len(imp.Syntax) == 0 {
			continue
		}
		deps[path] = imp.Syntax
	}

	return &tag.Context{
		Local:    main.Syntax,
		Resolver: tag.NewResolver(main.Syntax, deps),
	}, main.Name, nil
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
