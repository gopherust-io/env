package tag

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

type Field struct {
	TypeExpr  ast.Expr
	FieldPath string
	GoType    string
	EnvKey    string
	Default   string
	Name      string
	SliceElem string
	MapValue  string
	Prefix    string
	Sep       string
	KvSep     string
	Layout    string
	MapKey    string
	Children  []Field
	Required  bool
	IsPointer bool
	IsSlice   bool
	IsMap     bool
	IsNested  bool
	Expand    bool
	Skip      bool
	Sensitive bool
	Unmarshal bool
}

func ParseField(f *ast.Field, prefix, pathPrefix string, ctx *Context) ([]Field, error) {
	if f.Names == nil {
		return nil, nil
	}

	tags := parseTags(f.Tag)
	var out []Field

	for _, name := range f.Names {
		if !name.IsExported() {
			continue
		}

		fieldPath := name.Name
		if pathPrefix != "" {
			fieldPath = pathPrefix + "." + name.Name
		}

		goType, typeExpr, isPtr, isSlice, isMap, sliceElem, mapKey, mapVal, err := typeInfo(f.Type)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", name.Name, err)
		}

		envKey := tags.Env
		if envKey == "-" {
			continue
		}

		fieldPrefix := prefix
		if tags.Prefix != "" {
			fieldPrefix = prefix + tags.Prefix
		}

		if envKey == "" && strings.HasPrefix(tags.EnvRaw, "prefix:") {
			fieldPrefix = prefix + strings.TrimPrefix(tags.EnvRaw, "prefix:")
			envKey = ""
		}

		if isNested(goType, isPtr, isSlice, isMap) && envKey == "" {
			nested, err := parseNestedType(ctx, typeExpr, fieldPrefix, fieldPath, name.Name)
			if err != nil {
				return nil, err
			}
			if len(nested) > 0 {
				out = append(out, nested...)
				continue
			}
		}

		if envKey == "" {
			envKey = toSnakeUpper(name.Name)
		}

		unmarshal := ctx.hasUnmarshal(typeExpr)

		out = append(out, Field{
			Name:      name.Name,
			GoType:    goType,
			TypeExpr:  typeExpr,
			EnvKey:    envKey,
			Default:   tags.Default,
			Required:  tags.Required,
			Sensitive: tags.Sensitive,
			Skip:      false,
			Prefix:    fieldPrefix,
			Sep:       tags.Sep,
			KvSep:     tags.KvSep,
			Layout:    tags.Layout,
			Expand:    tags.Expand,
			IsNested:  false,
			IsPointer: isPtr,
			IsSlice:   isSlice,
			IsMap:     isMap,
			SliceElem: sliceElem,
			MapKey:    mapKey,
			MapValue:  mapVal,
			FieldPath: fieldPath,
			Unmarshal: unmarshal,
		})
	}

	return out, nil
}

func parseNestedType(ctx *Context, typeExpr ast.Expr, prefix, pathPrefix, parentName string) ([]Field, error) {
	expr := typeExpr
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.Ident, *ast.SelectorExpr:
			st, err := ctx.resolveStruct(expr)
			if err != nil {
				return nil, fmt.Errorf("field %s: nested type %s: %w", parentName, typeString(typeExpr), err)
			}
			if st == nil {
				return nil, fmt.Errorf("field %s: nested type %s not found", parentName, typeString(typeExpr))
			}
			var out []Field
			for _, f := range st.Fields.List {
				nested, err := ParseField(f, prefix, pathPrefix, ctx)
				if err != nil {
					return nil, err
				}
				out = append(out, nested...)
			}
			return []Field{{
				Name:      parentName,
				GoType:    typeString(typeExpr),
				TypeExpr:  typeExpr,
				Prefix:    prefix,
				FieldPath: pathPrefix,
				IsNested:  true,
				Children:  out,
			}}, nil
		case *ast.StructType:
			var out []Field
			for _, f := range t.Fields.List {
				nested, err := ParseField(f, prefix, pathPrefix, ctx)
				if err != nil {
					return nil, err
				}
				out = append(out, nested...)
			}
			return []Field{{
				Name:      parentName,
				GoType:    typeString(typeExpr),
				TypeExpr:  typeExpr,
				Prefix:    prefix,
				FieldPath: pathPrefix,
				IsNested:  true,
				Children:  out,
			}}, nil
		default:
			return nil, nil
		}
	}
}

type parsedTags struct {
	Env       string
	EnvRaw    string
	Default   string
	Prefix    string
	Sep       string
	KvSep     string
	Layout    string
	Required  bool
	Sensitive bool
	Expand    bool
}

func parseTags(tagLit *ast.BasicLit) parsedTags {
	if tagLit == nil {
		return parsedTags{}
	}
	val, err := strconv.Unquote(tagLit.Value)
	if err != nil {
		return parsedTags{}
	}

	var tags parsedTags
	for val != "" {
		key, rest := splitTag(val)
		val = rest
		if key == "" {
			break
		}
		if !strings.Contains(key, ":") {
			switch key {
			case "required":
				tags.Required = true
			case "sensitive":
				tags.Sensitive = true
			case "expand":
				tags.Expand = true
			}
			continue
		}
		k, v, _ := strings.Cut(key, ":")
		v = strings.Trim(v, `"`)
		switch k {
		case "env":
			tags.EnvRaw = v
			if p, ok := strings.CutPrefix(v, "prefix:"); ok {
				tags.Prefix = p
			} else {
				tags.Env = v
			}
		case "default":
			tags.Default = v
		case "prefix":
			tags.Prefix = v
		case "sep":
			tags.Sep = v
		case "kvsep":
			tags.KvSep = v
		case "layout":
			tags.Layout = v
		case "required":
			tags.Required = v != "false"
		case "sensitive":
			tags.Sensitive = v != "false"
		case "expand":
			tags.Expand = v != "false"
		}
	}
	return tags
}

func splitTag(s string) (key, rest string) {
	if s == "" {
		return "", ""
	}
	if s[0] == '"' {
		end := strings.Index(s[1:], `"`)
		if end < 0 {
			return s, ""
		}
		end++
		rest = strings.TrimSpace(s[end+1:])
		return s[:end+1], rest
	}
	if before, after, ok := strings.Cut(s, " "); ok {
		return strings.TrimSpace(before), strings.TrimSpace(after)
	}
	return strings.TrimSpace(s), ""
}

func typeInfo(expr ast.Expr) (goType string, typeExpr ast.Expr, isPtr, isSlice, isMap bool, sliceElem, mapKey, mapVal string, err error) {
	goType = typeString(expr)

	switch t := expr.(type) {
	case *ast.StarExpr:
		inner, _, _, _, _, _, _, _, err := typeInfo(t.X)
		if err != nil {
			return "", nil, false, false, false, "", "", "", err
		}
		return "*" + inner, expr, true, false, false, "", "", "", nil
	case *ast.ArrayType:
		elem := typeString(t.Elt)
		return "[]" + elem, expr, false, true, false, elem, "", "", nil
	case *ast.MapType:
		key := typeString(t.Key)
		val := typeString(t.Value)
		return "map[" + key + "]" + val, expr, false, false, true, "", key, val, nil
	default:
		return goType, expr, false, false, false, "", "", "", nil
	}
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.StructType:
		return "struct{}"
	default:
		return "interface{}"
	}
}

func isNested(goType string, isPtr, isSlice, isMap bool) bool {
	if isSlice || isMap {
		return false
	}
	if isPtr {
		goType = strings.TrimPrefix(goType, "*")
	}
	switch goType {
	case "string", "bool", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "time.Duration", "time.Time":
		return false
	default:
		return !strings.HasPrefix(goType, "[]") && !strings.HasPrefix(goType, "map[")
	}
}

func hasUnmarshalEnv(files []*ast.File, expr ast.Expr) bool {
	typeName := baseTypeName(expr)
	if typeName == "" {
		return false
	}
	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name.Name != "UnmarshalEnv" {
				continue
			}
			recvType := baseTypeName(fn.Recv.List[0].Type)
			if recvType == typeName {
				return true
			}
		}
	}
	return false
}

func baseTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return baseTypeName(t.X)
	default:
		return ""
	}
}

func toSnakeUpper(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func FindStruct(files []*ast.File, typeName string) (*ast.StructType, string, error) {
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
				if !ok {
					return nil, "", fmt.Errorf("type %s is not a struct", typeName)
				}
				return st, f.Name.Name, nil
			}
		}
	}
	return nil, "", fmt.Errorf("struct %s not found", typeName)
}

func ResolveStructType(files []*ast.File, expr ast.Expr) (*ast.StructType, error) {
	return resolveStructType(files, expr)
}

func CollectFields(ctx *Context, st *ast.StructType, pathPrefix string) ([]Field, error) {
	var out []Field
	for _, f := range st.Fields.List {
		fields, err := ParseField(f, "", pathPrefix, ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, fields...)
	}
	return out, nil
}
