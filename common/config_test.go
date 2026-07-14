package common

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// mapParser is a minimal koanf.Parser used to exercise the file-loading layer
// without pulling in a concrete format parser dependency.
type mapParser struct {
	out map[string]any
	err error
}

func (p mapParser) Unmarshal([]byte) (map[string]any, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.out, nil
}

func (p mapParser) Marshal(map[string]any) ([]byte, error) {
	return nil, nil
}

func TestLoadKoanfLayers_NilKoanf(t *testing.T) {
	if _, err := LoadKoanfLayers(nil, nil, nil, nil, nil, nil, ""); err == nil {
		t.Fatalf("expected error for nil koanf instance")
	}
}

func TestLoadKoanfLayers_Defaults(t *testing.T) {
	k := koanf.New(".")
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	usedFile, err := LoadKoanfLayers(
		k,
		map[string]any{"a": "default_a", "b": "default_b"},
		nil,
		nil,
		fs,
		nil,
		"defaults",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usedFile != "" {
		t.Fatalf("usedFile = %q, want empty", usedFile)
	}
	if k.String("a") != "default_a" {
		t.Fatalf("a = %q, want %q", k.String("a"), "default_a")
	}
	if k.String("b") != "default_b" {
		t.Fatalf("b = %q, want %q", k.String("b"), "default_b")
	}
}

func TestLoadKoanfLayers_FlagsOverrideDefaults(t *testing.T) {
	k := koanf.New(".")
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("b", "", "")
	if err := fs.Set("b", "flagval"); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}
	_, err := LoadKoanfLayers(
		k,
		map[string]any{"a": "default_a", "b": "default_b"},
		nil,
		nil,
		fs,
		nil,
		"flags",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k.String("a") != "default_a" {
		t.Fatalf("a = %q, want default_a (untouched)", k.String("a"))
	}
	if k.String("b") != "flagval" {
		t.Fatalf("b = %q, want flagval (flag override)", k.String("b"))
	}
}

func TestLoadKoanfLayers_EnvOverridesDefaults(t *testing.T) {
	t.Setenv("AGELANSERVER_ENVTEST_C", "envval")
	k := koanf.New(".")
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, err := LoadKoanfLayers(
		k,
		map[string]any{"C": "default_c"},
		nil,
		nil,
		fs,
		nil,
		"envtest",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k.String("C") != "envval" {
		t.Fatalf("C = %q, want envval (env override)", k.String("C"))
	}
}

func TestLoadKoanfLayers_FileLoaded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.conf")
	if err := os.WriteFile(path, []byte("ignored"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	k := koanf.New(".")
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	usedFile, err := LoadKoanfLayers(
		k,
		nil,
		[]string{"", filepath.Join(dir, "missing.conf"), path},
		mapParser{out: map[string]any{"fromfile": "x"}},
		fs,
		nil,
		"file",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usedFile != path {
		t.Fatalf("usedFile = %q, want %q", usedFile, path)
	}
	if k.String("fromfile") != "x" {
		t.Fatalf("fromfile = %q, want x", k.String("fromfile"))
	}
}

func TestLoadKoanfLayers_FileParseErrorWrapped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.conf")
	if err := os.WriteFile(path, []byte("bad"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	parseErr := errors.New("boom")
	k := koanf.New(".")
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	_, err := LoadKoanfLayers(
		k,
		nil,
		[]string{path},
		mapParser{err: parseErr},
		fs,
		nil,
		"file",
	)
	if err == nil {
		t.Fatalf("expected error")
	}
	var fileErr *KoanfFileLoadError
	if !errors.As(err, &fileErr) {
		t.Fatalf("error is not *KoanfFileLoadError: %T", err)
	}
	if fileErr.Path != path {
		t.Fatalf("fileErr.Path = %q, want %q", fileErr.Path, path)
	}
	if !errors.Is(err, parseErr) {
		t.Fatalf("expected wrapped error to match parseErr via errors.Is")
	}
}
