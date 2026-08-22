package battleServer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Regression: a single unparseable file (e.g. notes.txt) in the folder used to
// poison the named-return err of Configs, making valid configs unusable.
func TestConfigsIgnoresForeignFiles(t *testing.T) {
	gameId := "unit-" + strings.TrimPrefix(filepath.Base(t.TempDir()), "Test")
	dir := Folder(gameId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(dir)) })

	validToml := "Region = \"region-1\"\nIPv4 = \"127.0.0.1\"\nBsPort = 1\nWebSocketPort = 2\n"
	if err := os.WriteFile(filepath.Join(dir, Name(0)), []byte(validToml), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("unrelated junk"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	configs, err := Configs(gameId, false, true)
	if err != nil {
		t.Fatalf("foreign file must not fail the listing: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("configs = %d, want 1", len(configs))
	}
	if configs[0].index != 0 {
		t.Fatalf("index = %d, want 0", configs[0].index)
	}
	if configs[0].Region != "region-1" || configs[0].IPv4 != "127.0.0.1" || configs[0].BsPort != 1 || configs[0].WebSocketPort != 2 {
		t.Fatalf("parsed config mismatch: %+v", configs[0].Base)
	}
}

func TestConfigsMissingFolderIsNotAnError(t *testing.T) {
	configs, err := Configs("unit-never-created-xyz", false, true)
	if err != nil {
		t.Fatalf("missing folder must return nil error, got %v", err)
	}
	if len(configs) != 0 {
		t.Fatalf("configs = %d, want 0", len(configs))
	}
}

func TestConfigsBrokenTomlFailsWithFileName(t *testing.T) {
	gameId := "unit-broken"
	dir := Folder(gameId)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.WriteFile(filepath.Join(dir, Name(7)), []byte("%%%% not toml %%"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Configs(gameId, false, true)
	if err == nil {
		t.Fatal("expected parse error for broken toml")
	}
	if !strings.Contains(err.Error(), Name(7)) {
		t.Fatalf("error must mention the offending file: %v", err)
	}
}

func TestConfigsMissingFolderErrorIsNilEvenWithOnlyValid(t *testing.T) {
	// onlyValid=true must not change the missing-folder behavior.
	configs, err := Configs("unit-never-created-abc", true, false)
	if err != nil || len(configs) != 0 {
		t.Fatalf("got %v, %d", err, len(configs))
	}
}

func TestParseFileNameRejectsNonNumeric(t *testing.T) {
	if _, err := ParseFileName("config.toml"); err == nil {
		t.Fatal("expected error for alphabetic name")
	}
	if _, err := ParseFileName(""); err == nil {
		t.Fatal("expected error for empty name")
	}
}
