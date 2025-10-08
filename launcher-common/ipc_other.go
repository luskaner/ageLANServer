//go:build !windows

package launcher_common

import (
	"os"
	"path/filepath"
)

func ConfigAdminIpcPath() string {
	return filepath.Join(os.TempDir(), configAdminIpcName)
}
