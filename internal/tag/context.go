package tag

import (
	"go/ast"
)

// Context holds package-local files and optional cross-package resolution.
type Context struct {
	Resolver *Resolver
	Local    []*ast.File
}

func (c *Context) localFiles() []*ast.File {
	if c == nil || len(c.Local) == 0 {
		return nil
	}
	return c.Local
}

func (c *Context) resolveStruct(expr ast.Expr) (*ast.StructType, error) {
	if c != nil && c.Resolver != nil {
		return c.Resolver.ResolveStructType(expr)
	}
	return resolveStructType(c.localFiles(), expr)
}

func (c *Context) hasUnmarshal(expr ast.Expr) bool {
	if c != nil && c.Resolver != nil {
		return c.Resolver.HasUnmarshalEnv(expr)
	}
	return hasUnmarshalEnv(c.localFiles(), expr)
}
