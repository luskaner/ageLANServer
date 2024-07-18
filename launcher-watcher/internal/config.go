package internal

import (
	"common"
	launcherCommon "launcher-common"
	"launcher-common/executor"
	"log"
)

func RunConfig(revertFlags []string) {
	args := []string{launcherCommon.ConfigRevertCmd}
	args = append(args, revertFlags...)
	result := executor.ExecOptions{File: common.GetExeFileName(common.LauncherConfig), ExitCode: true, Wait: true, Args: args}.Exec()
	if !result.Success() {
		log.Println("Failed to run config.")
	}
}
