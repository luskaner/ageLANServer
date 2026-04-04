package admin

import (
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/luskaner/ageLANServer/common/logger"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
)

func postAgentStart(_ uint32, _ string) bool { return true }

func DialIPC() (net.Conn, error) {
	path := commonIpc.Path()
	commonLogger.Printf("Using %s\n", path)
	return winio.DialPipe(path, nil)
}
