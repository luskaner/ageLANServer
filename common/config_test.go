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

// The env provider strips the uppercase prefix, converts '_' to '.' and
// PRESERVES the case of the rest, so 'PREFIX_Ports_Bs' maps to 'Ports.Bs'.
func TestEnvProviderPreservesKeyCase(t *testing.T) {
	t.Setenv("AGELANSERVER_UNITTEST_Foo_Bar", "baz")
	t.Setenv("AGELANSERVER_UNITTEST_List", "a b c")

	k := koanf.New(".")
	if err := k.Load(koanfEnvProvider(Name+"_unittest"), nil); err != nil {
		t.Fatalf("load env: %v", err)
	}
	if got := k.String("Foo.Bar"); got != "baz" {
		t.Fatalf("Foo.Bar = %q, want %q", got, "baz")
	}
	if got := k.Strings("List"); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("List = %v, want [a b c]", got)
	}
}

// Regression for documented precedence defaults < env < file < flags.
func TestLoadKoanfLayersPrecedence(t *testing.T) {
	file := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(file, []byte(`{"Key":"from-file","onlyFile":"F"}`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGELANSERVER_UNITTEST_Key", "from-env")
	t.Setenv("AGELANSERVER_UNITTEST_OnlyEnv", "E")

	newKoanf := func() *koanf.Koanf { return koanf.New(".") }
	defaults := map[string]any{"Key": "from-defaults", "onlyDefault": "D"}

	// defaults < env
	k := newKoanf()
	if _, err := LoadKoanfLayers(k, defaults, nil, jsonParser{}, pflag.NewFlagSet("t", pflag.ContinueOnError), nil, "unittest"); err != nil {
		t.Fatal(err)
	}
	if got := k.String("Key"); got != "from-env" {
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
	if got := k.String("Key"); got != "from-file" {
		t.Fatalf("file must beat env, got %q", got)
	}
	if got := k.String("OnlyEnv"); got != "E" {
		t.Fatalf("env value lost when file present: %q", got)
	}

	// file < flags
	k = newKoanf()
	fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
	fs.String("Key", "", "")
	if err := fs.Set("Key", "from-flag"); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadKoanfLayers(k, defaults, []string{file}, jsonParser{}, fs, nil, "unittest"); err != nil {
		t.Fatal(err)
	}
	if got := k.String("Key"); got != "from-flag" {
		t.Fatalf("flags must beat file, got %q", got)
	}
	if got := k.String("onlyFile"); got != "F" {
		t.Fatalf("file value lost: %q", got)
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
