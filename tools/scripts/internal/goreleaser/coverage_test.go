package goreleaser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateSuccess(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)
	// Generate should create .goreleaser.yaml
	if err := Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if _, err := os.Stat(".goreleaser.yaml"); err != nil {
		t.Fatalf("goreleaser.yaml not created: %v", err)
	}
}

func TestGenerateCurrentExcludesWindows7Builds(t *testing.T) {
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := GenerateCurrent(); err != nil {
		t.Fatalf("GenerateCurrent failed: %v", err)
	}
	data, err := os.ReadFile(".goreleaser.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var project struct {
		Builds []struct {
			ID string `yaml:"id"`
		} `yaml:"builds"`
		UniversalBinaries []struct {
			ID string `yaml:"id"`
		} `yaml:"universal_binaries"`
	}
	if err := yaml.Unmarshal(data, &project); err != nil {
		t.Fatal(err)
	}
	hasWindows10Amd64 := false
	for _, build := range project.Builds {
		if strings.Contains(build.ID, "_win7_") {
			t.Errorf("current build config includes legacy Windows target %q", build.ID)
		}
		if build.ID == "server_win10_amd64" {
			hasWindows10Amd64 = true
		}
	}
	if !hasWindows10Amd64 {
		t.Error("current build config does not include Windows 10 amd64")
	}
	if len(project.UniversalBinaries) > 0 {
		t.Error("current build config includes universal binaries for other platforms")
	}
}

func TestAddSrcFile(t *testing.T) {
	archive := NewArchive("test", Targets64, nil)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	os.WriteFile(src, []byte("hello"), 0644)
	archive.AddSrcFile(src)
}

func TestAddSrcDstFileWithMode(t *testing.T) {
	archive := NewArchive("test", Targets64, nil)
	archive.AddSrcDstFileWithMode("src.txt", "dst.txt", 0644)
	// Ensure no panic
}

func TestAddScriptFilesAndConfig(t *testing.T) {
	archive := NewArchive("test", Targets64, nil)
	archive.AddScriptFiles("bin", NewTemplate[FileData]("{{.BaseOS}}/{{.Name}}"), nil, nil, false, false)
	archive.AddConfigFiles("config", NewTemplate[FileData]("config.toml"), true)
}

func TestAddAuxiliarBinary(t *testing.T) {
	archive := NewArchive("test", Targets64, nil)
	bin := NewBinary("./test", Targets64)
	archive.AddMainBinary(bin)
	aux := NewBinary("./aux", Targets64)
	archive.AddAuxiliarBinary(aux)
}

func TestCloneWithFilesPrefix(t *testing.T) {
	archive := NewArchive("test", Targets64, nil)
	archive.AddSrcDstFile("src.txt", "dst.txt")
	clone := archive.CloneWithFilesPrefix("prefix")
	if clone == nil {
		t.Fatal("clone nil")
	}
}

func TestNameAndGoarch(t *testing.T) {
	if OSWindowsModern.Name() == "" {
		t.Fatal("name empty")
	}
	if OSWindowsModern.Goos() == "" {
		t.Fatal("goos empty")
	}
	if ArchAmd64.Name() == "" {
		t.Fatal("arch name empty")
	}
	if ArchAmd64.Goarch() == "" {
		t.Fatal("goarch empty")
	}
}

func TestGenerateConfig(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)
	archive := NewArchive("test", Targets64, nil)
	archive.AddMainBinary(NewBinary("./test", Targets64))
	if err := GenerateConfig(archive); err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}
	if _, err := os.Stat(".goreleaser.yaml"); err != nil {
		t.Fatalf("not created: %v", err)
	}
}

func TestOverrideWindowsName(t *testing.T) {
	name := overrideWindowsName("full", OSWindowsModern, ArchArm64)
	if name != "win11" {
		t.Fatalf("got %q want win11", name)
	}
	name2 := overrideWindowsName("other", OSWindowsModern, ArchAmd64)
	if name2 != OSWindowsModern.Name() {
		t.Fatalf("got %q", name2)
	}
}
