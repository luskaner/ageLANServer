//go:build !windows

package watch

import (
	"fmt"
	"net"
	"os"
	"time"
)

func waitForProcess(pid uint32) bool {
	procPath := fmt.Sprintf("/proc/%d", pid)
	for {
		if _, err := os.Stat(procPath); os.IsNotExist(err) {
			return true
		}
		time.Sleep(10 * time.Second)
	}
}

func rebroadcastBattleServer(_ *int, _ []net.IP, _ int) {}
