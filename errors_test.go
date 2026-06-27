package env_test

import (
	"errors"
	"testing"

	"github.com/gopherust-io/env"
)

func TestErrorAggregation(t *testing.T) {
	var errs []env.FieldError
	env.AppendRequired(&errs, "Host", "DB_HOST")
	env.AppendParse(&errs, "Port", "PORT", "abc", errors.New("invalid syntax"))

	err := env.NewError(errs)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if msg == "" {
		t.Fatal("empty error message")
	}
	if err.Error() != msg {
		t.Fatal("expected lazy error message to be stable")
	}
}

func TestNewErrorNil(t *testing.T) {
	if err := env.NewError(nil); err != nil {
		t.Fatal("expected nil")
	}
}
