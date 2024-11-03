package internal

import (
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
)

func RunConfig(revertFlags []string) {
	args := []string{launcherCommon.ConfigRevertCmd}
	args = append(args, revertFlags...)
	exec.Options{File: common.GetExeFileName(true, common.LauncherConfig), ExitCode: true, Wait: true, Args: args}.Exec()
}
