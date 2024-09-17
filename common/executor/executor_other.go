//go:build !windows

package executor

import "os"

func IsAdmin() bool {
	return os.Geteuid() == 0
}
