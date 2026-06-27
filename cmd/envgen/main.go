package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/gopherust-io/env/internal/codegen"
	"github.com/gopherust-io/env/internal/tag"
)

func main() {
	dir := flag.String("dir", ".", "package directory")
	typeName := flag.String("type", "", "struct type name to generate loader for")
	output := flag.String("output", "", "output file (default: <snake>_env_gen.go)")
	pkgName := flag.String("package", "", "package name override")
	module := flag.String("module", "github.com/gopherust-io/env", "env module import path")
	list := flag.Bool("list", false, "list struct types in the package and exit")
	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fatal(err)
	}

	ctx, sourcePkg, err := loadPackageContext(absDir)
	if err != nil {
		fatal(err)
	}

	if *list {
		names := listStructs(ctx.Local)
		if len(names) == 0 {
			fmt.Fprintf(os.Stderr, "envgen: no structs found in %s\n", absDir)
			os.Exit(1)
		}
		for _, name := range names {
			fmt.Println(name)
		}
		return
	}

	if *typeName == "" {
		names := listStructs(ctx.Local)
		fmt.Fprintln(os.Stderr, "envgen: -type is required")
		if len(names) > 0 {
			fmt.Fprintf(os.Stderr, "envgen: structs in %s: %s\n", sourcePkg, strings.Join(names, ", "))
			fmt.Fprintf(os.Stderr, "envgen: example: envgen -dir %s -type %s\n", *dir, names[0])
		}
		os.Exit(2)
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
	fmt.Printf("envgen: add to your struct file:\n//go:generate envgen -type %s -output %s\n", *typeName, outFile)
}

func listStructs(files []*ast.File) []string {
	seen := make(map[string]struct{})
	for _, f := range files {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); ok && ts.Name.IsExported() {
					seen[ts.Name.Name] = struct{}{}
				}
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
