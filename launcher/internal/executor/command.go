package executor

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
)

func RunRevertCommand() (err error) {
	err = launcherCommon.RunRevertCommand()
	certStore.ReloadSystemCertificates()
	launcherCommon.ClearCache()
	return
}
