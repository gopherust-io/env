package reflectenv_test

import (
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/examples/basic"
	"github.com/gopherust-io/env/reflectenv"
)

func TestParseBasicConfig(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DEBUG", "true")
	t.Setenv("TIMEOUT", "5s")
	t.Setenv("STARTED", "2026-06-27")
	t.Setenv("NATS_URL", "nats://localhost:4222")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("TAGS", "a,b,c")
	t.Setenv("LABELS", "env:test,tier:1")
	env.ResetSnapshot()

	var cfg basic.Config
	if err := reflectenv.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9090 || !cfg.Debug || cfg.DB.Host != "localhost" {
		t.Fatalf("unexpected: %+v", cfg)
	}
}
