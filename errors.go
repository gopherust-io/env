package env

import (
	"strings"
	"sync"
)

// FieldError is a single field-level configuration error.
type FieldError struct {
	Field  string
	EnvKey string
	Op     string
	Value  string
	Err    error
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
	Fields []FieldError

	once sync.Once
	msg  string
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

func AppendParse(errs *[]FieldError, field, key, value string, parseErr error) {
	*errs = append(*errs, FieldError{
		Field:  field,
		EnvKey: key,
		Op:     "parse",
		Value:  value,
		Err:    parseErr,
	})
}
