package crosspkg_test

import (
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/examples/crosspkg"
)

func TestCrossPackageNested(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_PASSWORD", "secret")
	env.ResetSnapshot()

	cfg, err := crosspkg.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9090 || cfg.DB.Host != "db.local" || cfg.DB.Port != 5433 {
		t.Fatalf("unexpected: %+v", cfg)
	}
}
