package base

import (
	"fmt"

	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
)

type Executor interface {
	Do(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result)
	GameProcesses() (steamProcess bool, xboxProcess bool)
	fmt.Stringer
}
