//go:build !windows

package cmdUtils

import (
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"os"
)

func adminError(result *exec.Result) bool {
	return os.IsPermission(result.Err)
}
