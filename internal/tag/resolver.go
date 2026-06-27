package tag

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// Resolver maps import paths to package syntax trees for cross-package nested types.
type Resolver struct {
	byPath      map[string][]*ast.File
	byPkgName   map[string][]*ast.File
	importAlias map[string]string
	local       []*ast.File
}

func NewResolver(local []*ast.File, deps map[string][]*ast.File) *Resolver {
	r := &Resolver{
		local:     local,
		byPath:    deps,
		byPkgName: make(map[string][]*ast.File),
	}
	for path, files := range deps {
		if len(files) > 0 {
			r.byPkgName[files[0].Name.Name] = files
		}
		_ = path
	}
	r.importAlias = buildImportAlias(local)
	return r
}

func buildImportAlias(files []*ast.File) map[string]string {
	out := make(map[string]string)
	for _, f := range files {
		for _, imp := range f.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			name := path
			if imp.Name != nil {
				name = imp.Name.Name
			} else {
				if i := strings.LastIndex(path, "/"); i >= 0 {
					name = path[i+1:]
				}
			}
			out[name] = path
		}
	}
	return out
}

func (r *Resolver) filesForExpr(expr ast.Expr) []*ast.File {
	if r == nil {
		return nil
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return r.local
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			if path, ok := r.importAlias[id.Name]; ok {
				if files, ok := r.byPath[path]; ok {
					return files
				}
			}
			if files, ok := r.byPkgName[id.Name]; ok {
				return files
			}
		}
	}
	return r.local
}

func (r *Resolver) ResolveStructType(expr ast.Expr) (*ast.StructType, error) {
	if r == nil {
		return nil, fmt.Errorf("resolver is nil")
	}
	files := r.filesForExpr(expr)
	if files == nil {
		files = r.local
	}
	return resolveStructType(files, expr)
}

func (r *Resolver) HasUnmarshalEnv(expr ast.Expr) bool {
	if r == nil {
		return false
	}
	files := r.filesForExpr(expr)
	if files == nil {
		files = r.local
	}
	return hasUnmarshalEnv(files, expr)
}

func resolveStructType(files []*ast.File, expr ast.Expr) (*ast.StructType, error) {
	switch t := expr.(type) {
	case *ast.StructType:
		return t, nil
	case *ast.Ident:
		for _, f := range files {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name.Name != t.Name {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if ok {
						return st, nil
					}
				}
			}
		}
	case *ast.SelectorExpr:
		typeName := t.Sel.Name
		for _, f := range files {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name.Name != typeName {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if ok {
						return st, nil
					}
				}
			}
		}
	case *ast.StarExpr:
		return resolveStructType(files, t.X)
	}
	return nil, fmt.Errorf("unsupported struct type %T", expr)
}
