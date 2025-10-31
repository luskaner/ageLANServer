//go:build !windows

package internal

import (
	"net"
	"time"

	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

func preAgentStart() {
	if !executor.IsAdmin() {
		commonLogger.Println("Waiting up to 30s for 'agent' to start...")
	}
}

func postAgentStart(file string) {
	if !executor.IsAdmin() {
		for i := 0; i < 30; i++ {
			if _, proc, err := process.Process(file); err == nil && proc != nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
}

func DialIPC() (net.Conn, error) {
	path := launcherCommon.ConfigAdminIpcPath()
	commonLogger.Printf("Using unix:%s\n", path)
	return net.Dial("unix", path)
}
