package admin

import (
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/luskaner/ageLANServer/common/logger"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
)

func preAgentStart()          {}
func postAgentStart(_ string) {}

func DialIPC() (net.Conn, error) {
	path := commonIpc.Path()
	commonLogger.Printf("Using %s\n", path)
	return winio.DialPipe(path, nil)
}
