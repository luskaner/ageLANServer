package base

import (
	"testing"

	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
)

func TestStartUri(t *testing.T) {
	restore := commonExecutor.SetShellExecuteExCallFn(func(...uintptr) (uintptr, uintptr, error) { return 1, 0, nil })
	defer restore()
	var modified bool
	optionsFn := func(options commonExecutor.Options) {
		modified = true
		if !options.Shell {
			t.Error("Shell should be true")
		}
		if !options.SpecialFile {
			t.Error("SpecialFile should be true")
		}
		if !options.ShowWindow {
			t.Error("ShowWindow should be true")
		}
		if options.File != "https://example.com" {
			t.Errorf("File = %q, want %q", options.File, "https://example.com")
		}
	}
	result := StartUri("https://example.com", optionsFn)
	if !modified {
		t.Error("optionsFn was not called")
	}
	if result.Err != nil {
		t.Errorf("StartUri should succeed with mock, got %v", result.Err)
	}
}
