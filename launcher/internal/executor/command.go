package executor

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
)

func RunRevertCommand(optionsFn func(options exec.Options)) (err error) {
	err = launcherCommon.RunRevertCommand(optionsFn)
	certStore.ReloadSystemCertificates()
	common.ClearCache()
	return
}
