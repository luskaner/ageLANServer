package launcher_common

import (
	"net"

	"github.com/luskaner/ageLANServer/common"
)

const ConfigAdminIpcRevert byte = 0
const ConfigAdminIpcSetup byte = 1
const ConfigAdminIpcExit byte = 2
const configAdminIpcName = common.Name + `-launcher-config-admin-agent`

type (
	ConfigAdminIpcSetupCommand struct {
		CDN         bool
		IP          net.IP
		Certificate []byte
		GameId      string
	}
	ConfigAdminIpcRevertCommand struct {
		IPs         bool
		Certificate bool
	}
)
