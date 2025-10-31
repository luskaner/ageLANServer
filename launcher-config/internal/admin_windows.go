package internal

import (
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

func preAgentStart()          {}
func postAgentStart(_ string) {}

func DialIPC() (net.Conn, error) {
	path := launcherCommon.ConfigAdminIpcPath()
	commonLogger.Printf("Using %s\n", path)
	return winio.DialPipe(path, nil)
}
