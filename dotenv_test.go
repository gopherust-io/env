package env_test

import (
	"testing"

	"github.com/gopherust-io/env"
)

func TestParseDotEnvQuotesAndComments(t *testing.T) {
	data := []byte(`
# comment
URL=http://x.com#frag
KEY="foo#bar"
MIXED="hello'
EXPORT_OK=yes # trailing comment
export BAR=baz
QUOTED=""quoted""
PLAIN=noquotes
`)
	got, err := env.ParseDotEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"URL":       "http://x.com#frag",
		"KEY":       "foo#bar",
		"MIXED":     `"hello'`,
		"EXPORT_OK": "yes",
		"BAR":       "baz",
		"QUOTED":    `"quoted"`,
		"PLAIN":     "noquotes",
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("%s: got %q want %q (all=%v)", k, got[k], v, got)
		}
	}
}

func TestParseDotEnvExportPrefix(t *testing.T) {
	got, err := env.ParseDotEnv([]byte("export FOO=bar\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got["FOO"] != "bar" {
		t.Fatalf("got %#v", got)
	}
	if _, ok := got["export FOO"]; ok {
		t.Fatal("export prefix should be stripped from key")
	}
}
