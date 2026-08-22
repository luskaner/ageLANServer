package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

type jsonParser struct{}

func (jsonParser) Unmarshal(b []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	return m, err
}

func (jsonParser) Marshal(m map[string]interface{}) ([]byte, error) {
	return json.Marshal(m)
}

// Regression for the env layer being dead: keys must be lowercased so
// AGELANSERVER_<PREFIX>_FOO_BAR matches the koanf key foo.bar.
func TestEnvProviderLowercasesKeys(t *testing.T) {
	t.Setenv("AGELANSERVER_UNITTEST_FOO_BAR", "baz")
	t.Setenv("AGELANSERVER_UNITTEST_LIST", "a b c")

	k := koanf.New(".")
	if err := k.Load(koanfEnvProvider(Name+"_unittest"), nil); err != nil {
		t.Fatalf("load env: %v", err)
	}
	if got := k.String("foo.bar"); got != "baz" {
		t.Fatalf("foo.bar = %q, want %q (env layer must match lowercase keys)", got, "baz")
	}
	if got := k.Strings("list"); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("list = %v, want [a b c] (space-separated values become slices)", got)
	}
}

// Regression for documented precedence defaults < env < file < flags.
func TestLoadKoanfLayersPrecedence(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(file, []byte(`{"key":"from-file","onlyFile":"F"}`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGELANSERVER_UNITTEST_KEY", "from-env")
	t.Setenv("AGELANSERVER_UNITTEST_ONLYENV", "E")

	newKoanf := func() *koanf.Koanf { return koanf.New(".") }
	defaults := map[string]any{"key": "from-defaults", "onlyDefault": "D"}

	// defaults < env
	k := newKoanf()
	if _, err := LoadKoanfLayers(k, defaults, nil, jsonParser{}, pflag.NewFlagSet("t", pflag.ContinueOnError), nil, "unittest"); err != nil {
		t.Fatal(err)
	}
	if got := k.String("key"); got != "from-env" {
		t.Fatalf("env must beat defaults, got %q", got)
	}
	if got := k.String("onlyDefault"); got != "D" {
		t.Fatalf("defaults missing: %q", got)
	}

	// env < file (no flags involved)
	k = newKoanf()
	fsNone := pflag.NewFlagSet("none", pflag.ContinueOnError)
	if _, err := LoadKoanfLayers(k, defaults, []string{file}, jsonParser{}, fsNone, nil, "unittest"); err != nil {
		t.Fatal(err)
	}
	if got := k.String("key"); got != "from-file" {
		t.Fatalf("file must beat env, got %q", got)
	}
	// Env keys are always lowercase after the transform.
	if got := k.String("onlyenv"); got != "E" {
		t.Fatalf("env value lost when file present: %q", got)
	}

	// file < flags
	k = newKoanf()
	fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
	fs.String("key", "", "")
	if err := fs.Set("key", "from-flag"); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKoanfLayers(k, defaults, []string{file}, jsonParser{}, fs, nil, "unittest"); err != nil {
		t.Fatal(err)
	}
	if got := k.String("key"); got != "from-flag" {
		t.Fatalf("flags must beat file, got %q", got)
	}
}

func TestLoadKoanfLayersReturnsUsedFileAndMissingOk(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.json")
	k := koanf.New(".")
	used, err := LoadKoanfLayers(k, map[string]any{}, []string{missing}, jsonParser{}, pflag.NewFlagSet("t", pflag.ContinueOnError), nil, "unittest")
	if err != nil {
		t.Fatalf("missing file must be tolerated: %v", err)
	}
	if used != "" {
		t.Fatalf("used file = %q, want empty", used)
	}
}

func TestLoadKoanfLayersParseErrorWrapped(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	k := koanf.New(".")
	_, err := LoadKoanfLayers(k, nil, []string{bad}, jsonParser{}, pflag.NewFlagSet("t", pflag.ContinueOnError), nil, "unittest")
	if err == nil {
		t.Fatal("expected wrapped parse error")
	}
	fileErr, ok := err.(*KoanfFileLoadError)
	if !ok {
		t.Fatalf("error type = %T, want *KoanfFileLoadError", err)
	}
	if fileErr.Path != bad || fileErr.Err == nil {
		t.Fatalf("wrapped error incomplete: %+v", fileErr)
	}
}
