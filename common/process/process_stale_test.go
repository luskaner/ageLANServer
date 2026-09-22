package process

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/luskaner/ageLANServer/common"
)

// Regression: a stale (orphan) pid file used to leave a residual error
// (e.g. ESRCH) in the named return, so callers saw err != nil for the normal
// "nothing is running" case.
func TestProcessStalePidFileReturnsCleanState(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "ghost-process.exe")
	pidPaths := getPidPaths(exe)
	if len(pidPaths) == 0 {
		t.Fatal("getPidPaths returned no candidates")
	}
	stale := pidPaths[0]
	t.Cleanup(func() {
		for _, p := range pidPaths {
			_ = os.Remove(p)
		}
	})

	data := make([]byte, PidFileSize)
	binary.LittleEndian.PutUint64(data[0:8], uint64(4_000_000_000)) // dead PID
	binary.LittleEndian.PutUint64(data[8:16], 0)
	if err := os.WriteFile(stale, data, 0644); err != nil {
		t.Fatal(err)
	}

	pidPath, proc, err := Process(exe)
	if err != nil {
		t.Fatalf("stale pid file must not produce an error, got %v", err)
	}
	if proc != nil {
		t.Fatal("no live process expected")
	}
	if pidPath != stale {
		t.Fatalf("pidPath = %q, want first candidate %q", pidPath, stale)
	}
	if _, statErr := os.Stat(stale); !os.IsNotExist(statErr) {
		t.Fatal("stale pid file must be removed as orphan")
	}
}

func TestProcessNoPidFileAtAll(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "never-ran.exe")
	for _, p := range getPidPaths(exe) {
		_ = os.Remove(p)
	}
	pidPath, proc, err := Process(exe)
	if err != nil || proc != nil {
		t.Fatalf("got %v, %v; want clean not-running state", err, proc)
	}
	if pidPath == "" {
		t.Fatal("pidPath must still report where the lock would live")
	}
}

func TestGetPidPathsIncludesTempAndExeDir(t *testing.T) {
	exe := filepath.Join("some", "dir", "tool.exe")
	paths := getPidPaths(exe)
	wantTemp := filepath.Join(os.TempDir(), common.Name+"-tool.exe.pid")
	found := false
	for _, p := range paths {
		if p == wantTemp {
			found = true
		}
	}
	if !found {
		t.Fatalf("paths %v missing temp candidate %q", paths, wantTemp)
	}
	last := paths[len(paths)-1]
	if last != filepath.Join(filepath.Dir(exe), common.Name+"-tool.exe.pid") {
		t.Fatalf("exe-dir candidate = %q", last)
	}
}
