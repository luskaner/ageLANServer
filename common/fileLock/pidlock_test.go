package fileLock

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/process"
)

// Regression: openFile used to return (nil, nil) when another instance was
// running, and Lock(nil) produced confusing OS errors. It must now surface
// ErrAlreadyRunning.
func TestOpenFileAlreadyRunningSelf(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skip("cannot resolve executable")
	}
	name := common.Name + "-" + filepath.Base(exe) + ".pid"
	pidPath := filepath.Join(os.TempDir(), name)
	t.Cleanup(func() { _ = os.Remove(pidPath) })

	startTime, startErr := process.GetProcessStartTime(os.Getpid())
	if startErr != nil {
		// Without a real start time we cannot build a self-describing pid file.
		t.Skipf("GetProcessStartTime unavailable: %v", startErr)
	}
	data := make([]byte, process.PidFileSize)
	binary.LittleEndian.PutUint64(data[0:8], uint64(os.Getpid()))
	binary.LittleEndian.PutUint64(data[8:16], uint64(startTime))
	if err = os.WriteFile(pidPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	openErr, f := openFile()
	if !errors.Is(openErr, ErrAlreadyRunning) {
		t.Fatalf("openFile err = %v, want ErrAlreadyRunning", openErr)
	}
	if f != nil {
		t.Fatal("no file must be returned when already running")
	}

	lockErr := (&PidLock{}).Lock()
	if !errors.Is(lockErr, ErrAlreadyRunning) {
		t.Fatalf("Lock err = %v, want ErrAlreadyRunning", lockErr)
	}
}

func TestPidLockFullCycleAndIdempotentUnlock(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "unit.pid")

	f, err := os.OpenFile(pidPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	pl := &PidLock{}
	if err = pl.fileLock.Lock(f); err != nil {
		t.Fatalf("lock: %v", err)
	}
	if err = writePid(f); err != nil {
		t.Fatalf("writePid: %v", err)
	}

	// Contents round-trip.
	data, err := os.ReadFile(pidPath)
	if err != nil || len(data) != process.PidFileSize {
		t.Fatalf("pid file len = %d err = %v", len(data), err)
	}
	if got := binary.LittleEndian.Uint64(data[0:8]); int(got) != os.Getpid() {
		t.Fatalf("stored pid = %d", got)
	}

	if err = pl.Unlock(); err != nil {
		t.Fatalf("first Unlock: %v", err)
	}
	if _, statErr := os.Stat(pidPath); !os.IsNotExist(statErr) {
		t.Fatal("pid file must be removed on Unlock")
	}
	// Regression: a second Unlock used to operate on cleaned-up state.
	if err = pl.Unlock(); err != nil {
		t.Fatalf("second Unlock must be safe and nil, got %v", err)
	}
}
