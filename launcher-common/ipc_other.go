//go:build !windows

package launcher_common

import (
	"os"
	"path"
)

func ConfigAdminIpcPath() string {
	return path.Join(os.TempDir(), configAdminIpcName)
}
