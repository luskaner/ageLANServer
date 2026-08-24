package custom

import (
	"testing"
)

func TestString(t *testing.T) {
	e := Exec{Executable: "/some/path/app.exe"}
	if got := e.String(); got != "Custom Path" {
		t.Errorf("String() = %q, want %q", got, "Custom Path")
	}
}

func TestGameProcesses(t *testing.T) {
	e := Exec{Executable: "test"}
	steam, _, xbox := e.GameProcesses()
	if !steam {
		t.Error("steamProcess should be true")
	}
	// xbox depends on platform, just verify it doesn't panic
	_ = xbox
}
