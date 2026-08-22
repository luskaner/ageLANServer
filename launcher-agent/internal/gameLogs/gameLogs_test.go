package gameLogs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCopyPathToDirFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "log.txt")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	if !copyPathToDir(src, dstDir) {
		t.Fatal("copy reported failure")
	}
	data, err := os.ReadFile(filepath.Join(dstDir, "log.txt"))
	if err != nil || string(data) != "content" {
		t.Fatalf("copied content = %q err = %v", data, err)
	}
}

func TestCopyPathToDirDirectoryTree(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	tree := filepath.Join(srcDir, "session")
	if err := os.MkdirAll(filepath.Join(tree, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, "root.txt"), []byte("r"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, "sub", "leaf.txt"), []byte("l"), 0644); err != nil {
		t.Fatal(err)
	}

	if !copyPathToDir(tree, dstDir) {
		t.Fatal("copy reported failure")
	}
	for _, rel := range []string{"root.txt", filepath.Join("sub", "leaf.txt")} {
		data, err := os.ReadFile(filepath.Join(dstDir, "session", rel))
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		if len(data) == 0 {
			t.Fatalf("%s empty", rel)
		}
	}
}

func TestCopyPathToDirMissingSourceFails(t *testing.T) {
	if copyPathToDir(filepath.Join(t.TempDir(), "nope"), t.TempDir()) {
		t.Fatal("missing source must fail")
	}
}

func TestAddNewestPathPicksMostRecent(t *testing.T) {
	base := t.TempDir()
	old := filepath.Join(base, "old.txt")
	newer := filepath.Join(base, "new.txt")
	if err := os.WriteFile(old, []byte("o"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newer, []byte("n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Force distinct modification times (filesystems have 1s-2s granularity).
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(old, future, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	var paths []string
	addNewestPath(base, []string{old, newer}, func(info os.FileInfo) bool { return !info.IsDir() }, &paths)

	if len(paths) != 1 || paths[0] != newer {
		t.Fatalf("paths = %v, want newest %q", paths, newer)
	}
}

func TestCopyGameLogsUnknownGameIsNoop(t *testing.T) {
	CopyGameLogs("not-a-game", t.TempDir(), t.TempDir()) // must not panic
}
