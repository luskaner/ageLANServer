//go:build !windows

package ops

import (
	"os"

	"github.com/luskaner/ageServer/common/executor/exec"
)

func adminError(result *exec.Result) bool {
	return os.IsPermission(result.Err)
}
