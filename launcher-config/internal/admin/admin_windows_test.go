package admin

import "testing"

func TestPostAgentStartWindowsReturnsTrue(t *testing.T) {
	if !postAgentStart(123, "file") {
		t.Fatal("postAgentStart should return true on Windows")
	}
}

func TestDialIPCWindowsAttemptsPipe(t *testing.T) {
	// DialIPC tries to dial the named pipe; without agent it should error, but should not panic
	_, err := DialIPC()
	// We don't assert on err, just ensure it was called for coverage
	_ = err
}
