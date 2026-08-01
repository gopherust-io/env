package env

import "unsafe"

// IsEmpty reports whether s is empty.
func IsEmpty(s string) bool {
	return len(s) == 0
}

// StringToBytes returns a read-only view of s as a []byte without copying.
func StringToBytes(s string) []byte {
	if IsEmpty(s) {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString returns a string view of b without copying.
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
