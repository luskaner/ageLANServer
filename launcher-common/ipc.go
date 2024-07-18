package launcher_common

import (
	"common"
	"net"
)

const ConfigAdminIpcPipe = `\\.\pipe\` + common.Name + `-launcher-config-admin-agent`

const ConfigAdminIpcRevert byte = 0
const ConfigAdminIpcSetup byte = 1
const ConfigAdminIpcExit byte = 2

type (
	ConfigAdminIpcSetupCommand struct {
		IPs         []net.IP
		Certificate []byte
	}
	ConfigAdminIpcRevertCommand struct {
		IPs         bool
		Certificate bool
	}
)
