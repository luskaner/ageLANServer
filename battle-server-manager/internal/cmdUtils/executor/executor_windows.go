package executor

import (
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

var execWithOptions = func(_ string, options *exec.Options) exec.Result {
	return *options.Exec()
}
