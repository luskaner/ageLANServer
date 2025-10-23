package hosts

import (
	"github.com/luskaner/ageLANServer/common/executor/exec"
)

func FlushDns() (result *exec.Result) {
	result = exec.Options{File: "ipconfig", SpecialFile: true, UseWorkingPath: true, ExitCode: true, Wait: true, Args: []string{"/flushdns"}}.Exec()
	return
}
