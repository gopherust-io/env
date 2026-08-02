package env_test

import (
	"os"
	"strings"
	"testing"

	"github.com/gopherust-io/env"
)

func TestParseDotEnvExportTab(t *testing.T) {
	got, err := env.ParseDotEnv([]byte("export\tTABKEY=1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got["TABKEY"] != "1" {
		t.Fatalf("got %#v", got)
	}
	if _, ok := got["export\tTABKEY"]; ok {
		t.Fatal("export\\t prefix should be stripped")
	}
}

func TestParseDotEnvTooLarge(t *testing.T) {
	data := make([]byte, env.MaxDotEnvBytes+1)
	for i := range data {
		data[i] = 'a'
	}
	_, err := env.ParseDotEnv(data)
	if err == nil {
		t.Fatal("expected size error")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestParseDotEnvLongLineNoPartial(t *testing.T) {
	long := strings.Repeat("x", env.MaxDotEnvBytes)
	data := []byte("OK=1\nLONG=" + long + "\n")
	got, err := env.ParseDotEnv(data)
	if err == nil {
		t.Fatal("expected scanner/size error")
	}
	if got != nil {
		t.Fatalf("expected nil map on error, got %#v", got)
	}
}

func TestLoadDotEnvSetenvError(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/.env"
	// NUL in value is rejected by os.Setenv on most platforms
	if err := os.WriteFile(path, []byte("GOOD=ok\nBAD=a\x00b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("GOOD")
	_ = os.Unsetenv("BAD")

	err := env.LoadDotEnv(path)
	if err == nil {
		// Some platforms may accept NUL; skip if Setenv succeeded
		if _, ok := os.LookupEnv("BAD"); ok {
			t.Skip("platform accepts NUL in Setenv")
		}
		t.Fatal("expected Setenv error")
	}
	if !strings.Contains(err.Error(), "BAD") {
		t.Fatalf("expected BAD in error: %v", err)
	}
}
