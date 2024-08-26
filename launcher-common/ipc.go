package launcher_common

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"net"
	"os"
	"path"
)

const ConfigAdminIpcRevert byte = 0
const ConfigAdminIpcSetup byte = 1
const ConfigAdminIpcExit byte = 2

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

func ConfigAdminIpcName() string {
	return path.Join(os.TempDir(), common.Name+`-launcher-config-admin-agent`)
}
