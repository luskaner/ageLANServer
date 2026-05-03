package steam

import (
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
)

func (exec Exec) Do(_ []string, _ func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	// Should not get called
	return nil
}
