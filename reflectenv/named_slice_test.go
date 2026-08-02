package reflectenv_test

import (
	"testing"

	"github.com/gopherust-io/env"
	"github.com/gopherust-io/env/reflectenv"
)

type MyString string
type Tags []MyString

func TestParseNamedStringSlice(t *testing.T) {
	type Cfg struct {
		Tags Tags `env:"TAGS"`
	}
	var cfg Cfg
	err := reflectenv.ParseWithSnapshot(&cfg, env.FromMap(map[string]string{"TAGS": "a,b"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "a" || cfg.Tags[1] != "b" {
		t.Fatalf("Tags=%#v", cfg.Tags)
	}
}
