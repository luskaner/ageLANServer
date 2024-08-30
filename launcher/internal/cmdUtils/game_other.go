//go:build !windows

package cmdUtils

import (
	"github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
)

func adminError(result *exec.Result) bool {
	return result.ExitCode == 126
}
