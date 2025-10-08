package internal

import (
	"net"

	"github.com/Microsoft/go-winio"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

func preAgentStart()          {}
func postAgentStart(_ string) {}

func DialIPC() (net.Conn, error) {
	return winio.DialPipe(launcherCommon.ConfigAdminIpcPath(), nil)
}
