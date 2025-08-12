package launcher_common

import (
	"github.com/luskaner/ageLANServer/common"
	"net/netip"
)

const ConfigAdminIpcRevert byte = 0
const ConfigAdminIpcSetup byte = 1
const ConfigAdminIpcExit byte = 2
const configAdminIpcName = common.Name + `-launcher-config-admin-agent`

type (
	ConfigAdminIpcSetupCommand struct {
		CDN         bool
		IPAddr      netip.Addr
		Certificate []byte
	}
	ConfigAdminIpcRevertCommand struct {
		CDN         bool
		IPAddr      bool
		Certificate bool
	}
)
