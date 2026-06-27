package reflectenv

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gopherust-io/env"
)

// Parse reads environment variables into cfg using struct tags.
// It is a reflection-based alternative to codegen for prototyping and third-party types.
func Parse(cfg any) error {
	return ParseWithSnapshot(cfg, env.Snapshot())
}

// ParseWithSnapshot parses cfg using the given snapshot.
func ParseWithSnapshot(cfg any, snap *env.EnvSnapshot) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("reflectenv: cfg must be a pointer to struct")
	}
	var errs []env.FieldError
	parseStruct(v.Elem(), "", "", &errs, snap)
	return env.NewError(errs)
}

func parseStruct(v reflect.Value, prefix, pathPrefix string, errs *[]env.FieldError, snap *env.EnvSnapshot) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		tags := readFieldTags(sf)
		if tags.env == "-" {
			continue
		}

		fieldPath := sf.Name
		if pathPrefix != "" {
			fieldPath = pathPrefix + "." + sf.Name
		}

		fieldPrefix := prefix
		if tags.prefix != "" {
			fieldPrefix = prefix + tags.prefix
		}
		if tags.env != "" && strings.HasPrefix(tags.env, "prefix:") {
			fieldPrefix = prefix + strings.TrimPrefix(tags.env, "prefix:")
			tags.env = ""
		}

		fv := v.Field(i)
		if isNestedStruct(fv) && tags.env == "" {
			parseStruct(fv, fieldPrefix, fieldPath, errs, snap)
			continue
		}

		envKey := tags.env
		if envKey == "" {
			envKey = toSnakeUpper(sf.Name)
		}
		key := fieldPrefix + envKey
		loadField(fv, sf, fieldPath, key, tags, errs, snap)
	}
}

func loadField(v reflect.Value, sf reflect.StructField, fieldPath, key string, tags fieldTags, errs *[]env.FieldError, snap *env.EnvSnapshot) {
	raw, ok := snap.Lookup(key)
	if !ok || raw == "" {
		if tags.required && tags.defaultVal == "" {
			env.AppendRequired(errs, fieldPath, key)
			return
		}
		if tags.defaultVal != "" {
			raw = tags.defaultVal
		} else {
			return
		}
	}
	if tags.expand {
		raw = env.Expand(raw, snap)
	}

	if err := setValue(v, sf, tags, fieldPath, key, raw); err != nil {
		env.AppendParse(errs, fieldPath, key, raw, err)
	}
}

func setValue(v reflect.Value, sf reflect.StructField, tags fieldTags, fieldPath, key, raw string) error {
	if v.CanAddr() {
		if u, ok := v.Addr().Interface().(env.Unmarshaler); ok {
			return u.UnmarshalEnv(key, raw)
		}
	}

	if v.Type() == reflect.TypeFor[time.Duration]() {
		d, err := env.ParseDuration(raw)
		if err != nil {
			return err
		}
		v.SetInt(int64(d))
		return nil
	}
	if v.Type() == reflect.TypeFor[time.Time]() {
		layout := sf.Tag.Get("layout")
		if layout == "" {
			layout = time.RFC3339
		}
		tm, err := env.ParseTime(raw, layout)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(tm))
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
	case reflect.Bool:
		b, err := env.ParseBool(raw)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := parseIntKind(raw, v.Kind())
		if err != nil {
			return err
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := parseUintKind(raw, v.Kind())
		if err != nil {
			return err
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := parseFloatKind(raw, v.Kind())
		if err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String {
			sep := tags.sep
			if sep == "" {
				sep = ","
			}
			sl, err := env.ParseStringSlice(raw, sep)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(sl))
			return nil
		}
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Int {
			sep := tags.sep
			if sep == "" {
				sep = ","
			}
			sl, err := env.ParseIntSlice(raw, sep)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(sl))
			return nil
		}
		if v.Kind() == reflect.Map && v.Type().Key().Kind() == reflect.String && v.Type().Elem().Kind() == reflect.String {
			sep := tags.sep
			if sep == "" {
				sep = ","
			}
			kvSep := tags.kvSep
			if kvSep == "" {
				kvSep = ":"
			}
			m, err := env.ParseStringMap(raw, sep, kvSep)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(m))
			return nil
		}
		return fmt.Errorf("unsupported type %s", v.Type())
	}
	return nil
}

func isNestedStruct(v reflect.Value) bool {
	if v.Kind() != reflect.Struct {
		return false
	}
	if v.Type() == reflect.TypeFor[time.Time]() {
		return false
	}
	return true
}

type fieldTags struct {
	prefix     string
	defaultVal string
	env        string
	sep        string
	kvSep      string
	required   bool
	expand     bool
}

func readFieldTags(sf reflect.StructField) fieldTags {
	tags := parseFieldTags(sf.Tag.Get("env"))
	if p := sf.Tag.Get("prefix"); p != "" {
		tags.prefix = p
	}
	if !tags.required {
		if _, ok := sf.Tag.Lookup("required"); ok {
			v := sf.Tag.Get("required")
			tags.required = v == "" || v == "true"
		}
	}
	if !tags.expand {
		if _, ok := sf.Tag.Lookup("expand"); ok {
			v := sf.Tag.Get("expand")
			tags.expand = v == "" || v == "true"
		}
	}
	if s := sf.Tag.Get("sep"); s != "" {
		tags.sep = s
	}
	if s := sf.Tag.Get("kvsep"); s != "" {
		tags.kvSep = s
	}
	return tags
}

func parseFieldTags(tag string) fieldTags {
	var out fieldTags
	for tag != "" {
		key, rest := splitTag(tag)
		tag = rest
		if key == "" {
			break
		}
		if !strings.Contains(key, ":") {
			switch key {
			case "required":
				out.required = true
			case "expand":
				out.expand = true
			}
			continue
		}
		k, v, _ := strings.Cut(key, ":")
		v = strings.Trim(v, `"`)
		switch k {
		case "env":
			out.env = v
		case "default":
			out.defaultVal = v
		case "prefix":
			out.prefix = v
		case "required":
			out.required = v != "false"
		case "expand":
			out.expand = v != "false"
		}
	}
	return out
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
		return s[:end+1], strings.TrimSpace(s[end+1:])
	}
	if before, after, ok := strings.Cut(s, " "); ok {
		return strings.TrimSpace(before), strings.TrimSpace(after)
	}
	return strings.TrimSpace(s), ""
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

func parseIntKind(raw string, kind reflect.Kind) (int64, error) {
	switch kind {
	case reflect.Int:
		v, err := env.ParseInt(raw)
		return int64(v), err
	case reflect.Int8:
		v, err := env.ParseInt8(raw)
		return int64(v), err
	case reflect.Int16:
		v, err := env.ParseInt16(raw)
		return int64(v), err
	case reflect.Int32:
		v, err := env.ParseInt32(raw)
		return int64(v), err
	default:
		return env.ParseInt64(raw)
	}
}

func parseUintKind(raw string, kind reflect.Kind) (uint64, error) {
	switch kind {
	case reflect.Uint:
		v, err := env.ParseUint(raw)
		return uint64(v), err
	case reflect.Uint8:
		v, err := env.ParseUint8(raw)
		return uint64(v), err
	case reflect.Uint16:
		v, err := env.ParseUint16(raw)
		return uint64(v), err
	case reflect.Uint32:
		v, err := env.ParseUint32(raw)
		return uint64(v), err
	default:
		return env.ParseUint64(raw)
	}
}

func parseFloatKind(raw string, kind reflect.Kind) (float64, error) {
	if kind == reflect.Float32 {
		v, err := env.ParseFloat32(raw)
		return float64(v), err
	}
	return env.ParseFloat64(raw)
}
