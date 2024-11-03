package cmdUtils

import (
	"errors"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"golang.org/x/sys/windows"
)

func adminError(result *exec.Result) bool {
	return errors.Is(result.Err, windows.ERROR_ELEVATION_REQUIRED)
}
