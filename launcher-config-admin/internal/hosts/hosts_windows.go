package hosts

import (
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
)

func FlushDns() (result *exec.Result) {
	options := exec.Options{File: "ipconfig", SpecialFile: true, UseWorkingPath: true, ExitCode: true, Wait: true, Args: []string{"/flushdns"}}
	var suffix string
	if internal.SetUp {
		suffix = "_setup"
	} else {
		suffix = "_revert"
	}
	if err := internal.Logger.Buffer("ipconfig_flushdns"+suffix, func(writer io.Writer) {
		if writer != nil {
			options.Stdout = writer
			options.Stderr = writer
			commonLogger.Printf("run ipconfig: %s\n", options.String())
		}
		result = options.Exec()
	}); err != nil {
		result.ExitCode = common.ErrFileLog
		result.Err = err
	}
	return
}
