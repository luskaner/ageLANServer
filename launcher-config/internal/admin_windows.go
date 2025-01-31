package internal

import (
	"github.com/Microsoft/go-winio"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"net"
)

func preAgentStart()          {}
func postAgentStart(_ string) {}

func DialIPC() (net.Conn, error) {
	return winio.DialPipe(launcherCommon.ConfigAdminIpcPath(), nil)
}
