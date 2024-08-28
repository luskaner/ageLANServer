//go:build !windows && !darwin

package parentCheck

import (
	"fmt"
	"os"
)

func exePathFromPID(pid int) (string, error) {
	path := fmt.Sprintf("/proc/%d/exe", pid)
	exePath, err := os.Readlink(path)
	if err != nil {
		return "", err
	}
	return exePath, nil
}
