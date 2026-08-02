package reflectenv_test

import (
	"strings"
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/reflectenv"
)

func TestParseDefaultTag(t *testing.T) {
	type Cfg struct {
		Port int `env:"PORT" default:"8080"`
	}
	t.Setenv("PORT", "")
	env.ResetSnapshot()
	// empty PORT is treated as unset; default should apply
	var cfg Cfg
	if err := reflectenv.ParseWithSnapshot(&cfg, env.FromMap(map[string]string{})); err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port=%d", cfg.Port)
	}
}

func TestParseRequiredTruthy(t *testing.T) {
	type Cfg struct {
		Host string `env:"HOST" required:"1"`
	}
	var cfg Cfg
	err := reflectenv.ParseWithSnapshot(&cfg, env.FromMap(map[string]string{}))
	if err == nil {
		t.Fatal("expected required error")
	}
}

func TestParseSensitiveErrorRedacted(t *testing.T) {
	type Cfg struct {
		N int `env:"SECRET_N" sensitive:"true"`
	}
	secret := "not-an-int"
	err := reflectenv.ParseWithSnapshot(&Cfg{}, env.FromMap(map[string]string{"SECRET_N": secret}))
	if err == nil {
		t.Fatal("expected parse error")
	}
	if msg := err.Error(); strings.Contains(msg, secret) {
		t.Fatalf("leaked secret: %s", msg)
	}
}
