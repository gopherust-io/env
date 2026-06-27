package tag

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

type Field struct {
	Name       string
	GoType     string
	TypeExpr   ast.Expr
	EnvKey     string
	Default    string
	Required   bool
	Sensitive  bool
	Skip       bool
	Prefix     string
	Sep        string
	KvSep      string
	Children   []Field
	IsNested   bool
	IsPointer  bool
	IsSlice    bool
	IsMap      bool
	SliceElem  string
	MapKey     string
	MapValue   string
	FieldPath  string
	Unmarshal  bool
}

func ParseField(f *ast.Field, prefix, pathPrefix string, files []*ast.File) ([]Field, error) {
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
			nested, err := parseNestedType(files, typeExpr, fieldPrefix, fieldPath, name.Name)
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

		unmarshal := hasUnmarshalEnv(files, typeExpr)

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

func parseNestedType(files []*ast.File, typeExpr ast.Expr, prefix, pathPrefix, parentName string) ([]Field, error) {
	expr := typeExpr
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.Ident, *ast.SelectorExpr:
			st, err := ResolveStructType(files, expr)
			if err != nil {
				return nil, nil
			}
			var out []Field
			for _, f := range st.Fields.List {
				nested, err := ParseField(f, prefix, pathPrefix, files)
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
				nested, err := ParseField(f, prefix, pathPrefix, files)
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
	Required  bool
	Sensitive bool
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
			}
			continue
		}
		k, v, _ := strings.Cut(key, ":")
		v = strings.Trim(v, `"`)
		switch k {
		case "env":
			tags.EnvRaw = v
			if strings.HasPrefix(v, "prefix:") {
				tags.Prefix = strings.TrimPrefix(v, "prefix:")
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
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	}
	return strings.TrimSpace(s), ""
}

func typeInfo(expr ast.Expr) (goType string, typeExpr ast.Expr, isPtr, isSlice, isMap bool, sliceElem, mapKey, mapVal string, err error) {
	typeExpr = expr
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
		"float32", "float64", "time.Duration":
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
	case *ast.StarExpr:
		return ResolveStructType(files, t.X)
	}
	return nil, fmt.Errorf("unsupported struct type %T", expr)
}

func CollectFields(files []*ast.File, st *ast.StructType, pathPrefix string) ([]Field, error) {
	var out []Field
	for _, f := range st.Fields.List {
		fields, err := ParseField(f, "", pathPrefix, files)
		if err != nil {
			return nil, err
		}
		out = append(out, fields...)
	}
	return out, nil
}
