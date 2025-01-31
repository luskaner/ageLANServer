package internal

import (
	"fmt"
	"github.com/Microsoft/go-winio"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"net"
	"os/user"
)

func SetupIpcServer() (listener net.Listener, err error) {
	var u *user.User
	u, err = user.Current()
	if err != nil {
		return nil, err
	}
	pc := &winio.PipeConfig{
		InputBufferSize:    1_024,
		OutputBufferSize:   1,
		SecurityDescriptor: fmt.Sprintf("D:P(A;;GA;;;%s)", u.Uid),
		MessageMode:        true,
	}
	return winio.ListenPipe(launcherCommon.ConfigAdminIpcPath(), pc)
}

func RevertIpcServer() {}
