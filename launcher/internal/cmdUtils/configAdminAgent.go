package cmdUtils

import (
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/pflag"
)

func (c *Config) RunStopAgent() (result *exec.Result) {
	options := exec.Options{
		File:     executables.NativeFileName(false, executables.LauncherConfig),
		Wait:     true,
		Args:     cmd.FlagSetToArgs(pflag.NewFlagSet("stopAgent", pflag.ContinueOnError), true),
		ExitCode: true,
	}
	if buffErr := commonLogger.FileLogger.Buffer("stop_agent", func(writer io.Writer) {
		commonLogger.Println("run stop agent", options.String())
		if writer != nil {
			options.Stderr = writer
			options.Stdout = writer
		}
		result = options.Exec()
	}); buffErr != nil {
		result.Err = buffErr
		result.ExitCode = common.ErrFileLog
	}
	return
}
