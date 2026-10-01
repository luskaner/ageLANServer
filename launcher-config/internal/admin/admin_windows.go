package admin

import (
	"errors"
	"net"
	"os"

	"github.com/Microsoft/go-winio"
	"github.com/luskaner/ageLANServer/common/logger"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
	"golang.org/x/sys/windows"
)

func postAgentStart(_ uint32, _ string) bool { return true }

// isAccessDenied reports whether err is the privilege failure we get when a
// non-elevated process tries to terminate an elevated one.
func isAccessDenied(err error) bool {
	return errors.Is(err, windows.ERROR_ACCESS_DENIED) || errors.Is(err, os.ErrPermission)
}

func DialIPC() (net.Conn, error) {
	path := commonIpc.Path()
	commonLogger.Printf("Using %s\n", path)
	return winio.DialPipe(path, nil)
}
