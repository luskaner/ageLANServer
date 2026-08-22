package watch

import (
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// Helper process: re-executed by the tests to create real child processes
// that sleep for SLEEP_MS milliseconds and exit successfully.
func TestHelperProcessSleep(t *testing.T) {
	if os.Getenv("GO_WATCH_HELPER") != "1" {
		return
	}
	ms, _ := strconv.Atoi(os.Getenv("SLEEP_MS"))
	time.Sleep(time.Duration(ms) * time.Millisecond)
	os.Exit(0)
}

func spawnSleeper(t *testing.T, ms int) *os.Process {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessSleep$", "-test.v")
	cmd.Env = append(os.Environ(), "GO_WATCH_HELPER=1", "SLEEP_MS="+strconv.Itoa(ms))
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	return cmd.Process
}

func TestWaitForProcessesToExitWaitsForAll(t *testing.T) {
	fast := spawnSleeper(t, 50)
	slow := spawnSleeper(t, 500)

	start := time.Now()
	if !waitForProcessesToExit([]*os.Process{fast, slow}) {
		t.Fatal("expected successful wait")
	}
	elapsed := time.Since(start)
	if elapsed < 400*time.Millisecond {
		t.Fatalf("returned after %v; it did not wait for ALL processes (regression: single random pick)", elapsed)
	}
}

func TestWaitForProcessesToExitEmptyAndNil(t *testing.T) {
	if !waitForProcessesToExit(nil) {
		t.Fatal("empty list must succeed")
	}
	if !waitForProcessesToExit([]*os.Process{nil}) {
		t.Fatal("nil entries must be skipped")
	}
}

func TestSortedProcessNames(t *testing.T) {
	processes := map[string]*os.Process{
		"zebra.exe": nil,
		"alpha.exe": nil,
		"mid.exe":   nil,
	}
	names := sortedProcessNames(processes)
	want := []string{"alpha.exe", "mid.exe", "zebra.exe"}
	if len(names) != len(want) {
		t.Fatalf("names = %v", names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}
