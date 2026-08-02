package env

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

// FieldError is a single field-level configuration error.
type FieldError struct {
	Err       error
	Field     string
	EnvKey    string
	Op        string
	Value     string
	Sensitive bool
	_         [7]byte
}

func (e FieldError) Error() string {
	var b strings.Builder
	b.WriteString(e.Field)
	b.WriteString(" (")
	b.WriteString(e.EnvKey)
	b.WriteString("): ")
	b.WriteString(e.Op)
	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}
	return b.String()
}

// Error collects every field error from one parse pass.
type Error struct {
	msg    string
	Fields []FieldError
	once   sync.Once
}

func (e *Error) Error() string {
	e.once.Do(func() {
		if len(e.Fields) == 0 {
			e.msg = "env: configuration error"
			return
		}
		var b strings.Builder
		b.WriteString("env: ")
		for i, f := range e.Fields {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(f.Error())
		}
		e.msg = b.String()
	})
	return e.msg
}

// NewError returns nil when fields is empty.
func NewError(fields []FieldError) error {
	if len(fields) == 0 {
		return nil
	}
	return &Error{Fields: fields}
}

func AppendRequired(errs *[]FieldError, field, key string) {
	*errs = append(*errs, FieldError{
		Field:  field,
		EnvKey: key,
		Op:     "required",
	})
}

// AppendParse records a parse failure. Prefer AppendParseSensitive for sensitive fields.
func AppendParse(errs *[]FieldError, field, key, value string, parseErr error) {
	appendParse(errs, field, key, value, parseErr, false)
}

// AppendParseSensitive records a parse failure without retaining or printing the raw value.
func AppendParseSensitive(errs *[]FieldError, field, key, value string, parseErr error) {
	appendParse(errs, field, key, value, parseErr, true)
}

func appendParse(errs *[]FieldError, field, key, value string, parseErr error, sensitive bool) {
	fe := FieldError{
		Field:     field,
		EnvKey:    key,
		Op:        "parse",
		Value:     value,
		Err:       parseErr,
		Sensitive: sensitive,
	}
	if sensitive {
		fe.Value = SensitiveMask
		fe.Err = redactParseErr(parseErr, value)
	}
	*errs = append(*errs, fe)
}

func redactParseErr(parseErr error, value string) error {
	if parseErr == nil {
		return nil
	}
	msg := parseErr.Error()
	if value == "" {
		return errors.New(msg)
	}
	quoted := strconv.Quote(value)
	msg = strings.ReplaceAll(msg, quoted, strconv.Quote(SensitiveMask))
	if strings.Contains(msg, value) {
		msg = strings.ReplaceAll(msg, value, SensitiveMask)
	}
	// Opaque fallback when the value still appears (unusual encodings).
	if strings.Contains(msg, value) {
		return errors.New("invalid value")
	}
	return errors.New(msg)
}
