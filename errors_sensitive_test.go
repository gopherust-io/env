package env_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/gopherust-io/env"
)

func TestAppendParseSensitiveRedactsValue(t *testing.T) {
	secret := "supersecret"
	_, parseErr := strconv.ParseInt(secret, 10, 64)

	var errs []env.FieldError
	env.AppendParseSensitive(&errs, "Password", "DB_PASSWORD", secret, parseErr)

	if len(errs) != 1 {
		t.Fatalf("len=%d", len(errs))
	}
	fe := errs[0]
	if fe.Value != env.SensitiveMask {
		t.Fatalf("Value=%q", fe.Value)
	}
	if !fe.Sensitive {
		t.Fatal("expected Sensitive")
	}
	msg := fe.Error()
	if strings.Contains(msg, secret) {
		t.Fatalf("error leaked secret: %s", msg)
	}
	if !strings.Contains(msg, env.SensitiveMask) {
		t.Fatalf("expected mask in error: %s", msg)
	}

	agg := env.NewError(errs)
	if strings.Contains(agg.Error(), secret) {
		t.Fatalf("aggregated error leaked secret: %s", agg.Error())
	}
}

func TestAppendParseKeepsNonSensitiveValue(t *testing.T) {
	var errs []env.FieldError
	env.AppendParse(&errs, "Port", "PORT", "abc", errors.New(`strconv.ParseInt: parsing "abc": invalid syntax`))
	if errs[0].Value != "abc" {
		t.Fatalf("Value=%q", errs[0].Value)
	}
	if errs[0].Sensitive {
		t.Fatal("unexpected Sensitive")
	}
}

func TestAppendParseSensitiveRedactsUnquoted(t *testing.T) {
	secret := "s3cret-token"
	var errs []env.FieldError
	env.AppendParseSensitive(&errs, "Password", "DB_PASSWORD", secret, errors.New("bad: "+secret))
	msg := errs[0].Error()
	if strings.Contains(msg, secret) {
		t.Fatalf("leaked secret: %s", msg)
	}
	if !strings.Contains(msg, env.SensitiveMask) && !strings.Contains(msg, "invalid value") {
		t.Fatalf("expected mask or opaque error: %s", msg)
	}
}
