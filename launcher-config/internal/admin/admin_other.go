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

// postAgentStart has no pid to wait for on Windows, because runas is a Windows
// verb, so on the other platforms this polls instead. It used to loop with no
// bound and no failure path, so an agent that failed to start left config.exe
// hanging forever with no diagnostic.
const postAgentStartAttempts = 60

var (
	processWaitForProcessFn = process.WaitForProcess
	processFn               = process.Process
	sleepFn                 = time.Sleep
)

func postAgentStart(pid uint32, file string) (ok bool) {
	if executor.IsAdmin() {
		return true
	}
	for range postAgentStartAttempts {
		// WaitForProcess waits for exit, so a true here means the agent died
		// within the interval, i.e. it never came up. Fail fast on that rather
		// than burning the whole budget.
		if processWaitForProcessFn(&os.Process{Pid: int(pid)}, &waitInterval) {
			return false
		}
		if _, proc, err := processFn(file); err == nil && proc != nil {
			return true
		}
		sleepFn(time.Second)
	}
	return false
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
