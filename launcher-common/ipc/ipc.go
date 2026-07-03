package ipc

import (
	"net"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
)

const (
	Revert byte = iota
	Setup
	Exit
)

const name = common.Name + `-` + executables.LauncherConfigAdminAgent

type (
	SetupCommand struct {
		IP                     net.IP
		MacOsExclusiveMappings bool
		Certificate            []byte
		GameId                 string
	}
	RevertCommand struct {
		IPs         bool
		Certificate bool
	}
)
