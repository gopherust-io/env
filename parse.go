package env

import (
	"strconv"
	"strings"
	"time"
)

func ParseString(s string) (string, error) {
	return s, nil
}

func ParseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "y", "yes", "on":
		return true, nil
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	default:
		return strconv.ParseBool(s)
	}
}

func ParseInt(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, strconv.IntSize)
	return int(v), err
}

func ParseInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	return int8(v), err
}

func ParseInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

func ParseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ParseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, strconv.IntSize)
	return uint(v), err
}

func ParseUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}

func ParseUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

func ParseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

func ParseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func ParseFloat32(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

func ParseTime(s, layout string) (time.Time, error) {
	if layout == "" {
		layout = time.RFC3339
	}
	return time.Parse(layout, s)
}

func ParseStringSlice(s, sep string) ([]string, error) {
	if s == "" {
		return nil, nil
	}
	if sep == "" {
		sep = ","
	}
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

func ParseIntSlice(s, sep string) ([]int, error) {
	parts, err := ParseStringSlice(s, sep)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, nil
	}
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := ParseInt(p)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func ParseStringMap(s, sep, kvSep string) (map[string]string, error) {
	if s == "" {
		return nil, nil
	}
	if sep == "" {
		sep = ","
	}
	if kvSep == "" {
		kvSep = ":"
	}
	parts := strings.Split(s, sep)
	out := make(map[string]string, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, kvSep, 2)
		if len(kv) != 2 {
			return nil, strconv.ErrSyntax
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if key == "" {
			return nil, strconv.ErrSyntax
		}
		out[key] = val
	}
	return out, nil
}
