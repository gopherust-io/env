package minimal_test

import (
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/examples/minimal"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("DEBUG", "true")
	env.ResetSnapshot()

	cfg, err := minimal.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 3000 || !cfg.Debug || cfg.Host != "localhost" {
		t.Fatalf("unexpected: %+v", cfg)
	}
}
