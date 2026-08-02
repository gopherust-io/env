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

	"golang.org/x/mod/module"
	"golang.org/x/tools/go/packages"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/internal/codegen"
	"github.com/gopherust-io/env/internal/tag"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "doctor" {
		os.Exit(runDoctor(os.Args[2:]))
	}

	dir := flag.String("dir", ".", "package directory")
	typeName := flag.String("type", "", "struct type name to generate loader for")
	output := flag.String("output", "", "output file (default: <snake>_env_gen.go)")
	pkgName := flag.String("package", "", "package name override")
	modulePath := flag.String("module", "github.com/gopherust-io/env", "env module import path")
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

	if env.IsEmpty(*typeName) {
		names := listStructs(ctx.Local)
		fmt.Fprintln(os.Stderr, "envgen: -type is required")
		if len(names) > 0 {
			fmt.Fprintf(os.Stderr, "envgen: structs in %s: %s\n", absDir, strings.Join(names, ", "))
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
	if !env.IsEmpty(*pkgName) {
		outPkg = *pkgName
	}
	if !token.IsIdentifier(outPkg) {
		fatal(fmt.Errorf("invalid package name %q", outPkg))
	}
	if !token.IsIdentifier(*typeName) {
		fatal(fmt.Errorf("invalid type name %q", *typeName))
	}
	if err := module.CheckImportPath(*modulePath); err != nil {
		fatal(fmt.Errorf("invalid -module %q: %w", *modulePath, err))
	}

	outFile := *output
	if env.IsEmpty(outFile) {
		outFile = toSnake(*typeName) + "_env_gen.go"
	}

	outPath, err := resolveOutputPath(absDir, outFile)
	if err != nil {
		fatal(err)
	}

	src, err := codegen.Generate(codegen.Options{
		Package:  outPkg,
		TypeName: *typeName,
		Fields:   fields,
		Module:   *modulePath,
	})
	if err != nil {
		fatal(err)
	}

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

	// Avoid NeedDeps (full transitive syntax graph). Load the target package,
	// then lazily load syntax only for its direct imports.
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedSyntax | packages.NeedModule | packages.NeedFiles | packages.NeedImports,
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
	missing := make([]string, 0)
	for path, imp := range main.Imports {
		if imp == nil {
			continue
		}
		if len(imp.Syntax) > 0 {
			deps[path] = imp.Syntax
			continue
		}
		missing = append(missing, path)
	}
	if len(missing) > 0 {
		depCfg := &packages.Config{
			Mode:  packages.NeedName | packages.NeedSyntax | packages.NeedFiles,
			Dir:   absDir,
			Tests: false,
		}
		depPkgs, loadErr := packages.Load(depCfg, missing...)
		if loadErr != nil {
			return nil, "", loadErr
		}
		if packages.PrintErrors(depPkgs) > 0 {
			return nil, "", fmt.Errorf("import package load failed for nested types in %s", absDir)
		}
		for _, dp := range depPkgs {
			if len(dp.Syntax) == 0 {
				continue
			}
			deps[dp.PkgPath] = dp.Syntax
		}
	}

	return &tag.Context{
		Local:    main.Syntax,
		Resolver: tag.NewResolver(main.Syntax, deps),
	}, main.Name, nil
}

func resolveOutputPath(absDir, outFile string) (string, error) {
	if filepath.IsAbs(outFile) {
		return "", fmt.Errorf("-output must be relative to -dir, got absolute path %q", outFile)
	}
	outPath := filepath.Clean(filepath.Join(absDir, outFile))
	rel, err := filepath.Rel(absDir, outPath)
	if err != nil {
		return "", fmt.Errorf("-output %q escapes -dir: %w", outFile, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("-output %q escapes -dir %q", outFile, absDir)
	}
	return outPath, nil
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
