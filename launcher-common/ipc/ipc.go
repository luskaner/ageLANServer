package ipc

import (
	"net"

	"github.com/luskaner/ageLANServer/common"
)

const Revert byte = 0
const Setup byte = 1
const Exit byte = 2
const name = common.Name + `-launcher-config-admin-agent`

type (
	SetupCommand struct {
		IP          net.IP
		Certificate []byte
		GameId      string
	}
	RevertCommand struct {
		IPs         bool
		Certificate bool
	}
)
