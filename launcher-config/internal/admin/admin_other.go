//go:build !windows

package admin

import (
	"errors"
	"net"
	"os"

	"github.com/luskaner/ageLANServer/common/logger"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
	commonUi "github.com/luskaner/ageLANServer/launcher-common/ui"
)

// isAccessDenied reports whether err is the privilege failure we get when a
// non-elevated process tries to terminate an elevated one.
func isAccessDenied(err error) bool {
	return errors.Is(err, os.ErrPermission)
}

func DialIPC() (net.Conn, error) {
	path := commonIpc.Path()
	commonLogger.Println(commonUi.Detail("Using unix:%s", path))
	return net.Dial("unix", path)
}
