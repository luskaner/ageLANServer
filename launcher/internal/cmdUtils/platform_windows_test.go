package cmdUtils

import (
	"testing"

	"github.com/luskaner/ageLANServer/common/executor/exec"
	"golang.org/x/sys/windows"
)

// Split out of platform_test.go: ERROR_ELEVATION_REQUIRED is a Windows
// constant. Moved verbatim; only the location changed.
func TestAdminErrorWithElevationRequired(t *testing.T) {
	result := &exec.Result{Err: windows.ERROR_ELEVATION_REQUIRED}
	if !adminError(result) {
		t.Error("expected true for ERROR_ELEVATION_REQUIRED")
	}
}
