package basic_test

import (
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/examples/basic"
)

func TestLoadConfig(t *testing.T) {
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

	cfg, err := basic.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Port != 9090 || !cfg.Debug || cfg.DB.Host != "localhost" || cfg.DB.Port != 5433 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if len(cfg.Tags) != 3 || cfg.Labels["tier"] != "1" {
		t.Fatalf("unexpected tags/labels: %+v %+v", cfg.Tags, cfg.Labels)
	}
	if cfg.BaseURL != "nats://localhost:4222/api" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}

	masked := cfg.Masked()
	if masked.DB.Password != env.SensitiveMask {
		t.Fatalf("expected masked password, got %q", masked.DB.Password)
	}
}

func TestReloadConfig(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DB_HOST", "localhost")
	env.ResetSnapshot()

	cfg, err := basic.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("port = %d", cfg.Port)
	}

	t.Setenv("PORT", "9090")
	if err := basic.ReloadConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("after reload port = %d", cfg.Port)
	}
}

func TestRequiredField(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DB_HOST", "")
	env.ResetSnapshot()

	_, err := basic.LoadConfig()
	if err == nil {
		t.Fatal("expected required field error")
	}
}
