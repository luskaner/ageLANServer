//go:build !windows

package cmdUtils

import (
	"os"

	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func adminError(result *exec.Result) bool {
	return os.IsPermission(result.Err)
}
