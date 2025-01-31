//go:build !windows

package internal

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"net"
	"os"
)

func SetupIpcServer() (listener net.Listener, err error) {
	ipcPath := launcherCommon.ConfigAdminIpcPath()

	if err = os.Remove(ipcPath); err != nil && !os.IsNotExist(err) {
		return
	}

	listener, err = net.Listen("unix", ipcPath)

	if err != nil {
		return
	}

	err = os.Chmod(ipcPath, 0666)
	return
}

func RevertIpcServer() {
	_ = os.Remove(launcherCommon.ConfigAdminIpcPath())
}
