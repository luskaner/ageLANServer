//go:build !windows

package cmdUtils

import (
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"os"
)

func adminError(result *exec.Result) bool {
	return os.IsPermission(result.Err)
}
