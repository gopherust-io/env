// Package env is a codegen-first environment variable parser for Go.
// Use cmd/envgen to generate type-specific loaders with zero runtime reflection.
package env

// Unmarshaler parses a custom type from a raw environment value.
type Unmarshaler interface {
	UnmarshalEnv(key, value string) error
}

// SensitiveMask replaces sensitive fields in generated Masked() output.
const SensitiveMask = "***"
