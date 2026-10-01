//go:build !windows

package admin

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
	commonIpc "github.com/luskaner/ageLANServer/launcher-common/ipc"
)

var waitInterval = 100 * time.Millisecond

func postAgentStart(pid uint32, file string) (ok bool) {
	if executor.IsAdmin() {
		ok = true
	} else {
		for {
			if process.WaitForProcess(&os.Process{Pid: int(pid)}, &waitInterval) {
				break
			}
			if _, proc, err := process.Process(file); err == nil && proc != nil {
				ok = true
				break
			}
			time.Sleep(time.Second)
		}
	}
	return
}

// isAccessDenied reports whether err is the privilege failure we get when a
// non-elevated process tries to terminate an elevated one.
func isAccessDenied(err error) bool {
	return errors.Is(err, os.ErrPermission)
}

func DialIPC() (net.Conn, error) {
	path := commonIpc.Path()
	commonLogger.Printf("Using unix:%s\n", path)
	return net.Dial("unix", path)
}
