package executor

import (
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/certStore"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

func RunRevertCommand(out io.Writer, optionsFn func(options exec.Options)) (err error) {
	err = launcherCommon.RunRevertCommand(out, optionsFn)
	certStore.ReloadSystemCertificates()
	common.ClearDNSCache()
	return
}
