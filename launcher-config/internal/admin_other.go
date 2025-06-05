//go:build !windows

package internal

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"net"
	"time"
)

func preAgentStart() {
	if !executor.IsAdmin() {
		fmt.Println("Waiting up to 30s for 'agent' to start...")
	}
}

func postAgentStart(file string) {
	if !executor.IsAdmin() {
		for i := 0; i < 30; i++ {
			if _, proc, err := process.Process(file); err == nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
}

func DialIPC() (net.Conn, error) {
	return net.Dial("unix", launcherCommon.ConfigAdminIpcPath())
}
