package executor

import (
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func execWithOptions(_ string, options *exec.Options) exec.Result {
	return *options.Exec()
}
