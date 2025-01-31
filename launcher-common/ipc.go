package launcher_common

import (
	"github.com/luskaner/ageLANServer/common"
	"net"
)

const ConfigAdminIpcRevert byte = 0
const ConfigAdminIpcSetup byte = 1
const ConfigAdminIpcExit byte = 2
const configAdminIpcName = common.Name + `-launcher-config-admin-agent`

type (
	ConfigAdminIpcSetupCommand struct {
		CDN         bool
		IPs         []net.IP
		Certificate []byte
	}
	ConfigAdminIpcRevertCommand struct {
		CDN         bool
		IPs         bool
		Certificate bool
	}
)
