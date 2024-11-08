package watch

import (
	"github.com/luskaner/ageLANServer/common"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	launcherCommonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor"
	"time"
)

var processWaitInterval = 1 * time.Second

func waitUntilAnyProcessExist(names []string) (processesPID map[string]uint32) {
	for i := 0; i < int((1*time.Minute)/processWaitInterval); i++ {
		processesPID = commonProcess.ProcessesPID(names)
		if len(processesPID) > 0 {
			return
		}
		time.Sleep(processWaitInterval)
	}
	return
}

func Watch(gameId string, steamProcess bool, microsoftStoreProcess bool, serverExe string, broadcastBattleServer bool, revertArgs []string, revertCmd []string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if len(revertCmd) > 0 {
		defer func() {
			_ = launcherCommonExecutor.RunRevertCommand(revertCmd)
		}()
	}
	if len(revertArgs) > 0 {
		defer func() {
			internal.RunConfig(revertArgs)
		}()
	}
	processes := waitUntilAnyProcessExist(commonProcess.GameProcesses(gameId, steamProcess, microsoftStoreProcess))
	if len(processes) == 0 {
		*exitCode = internal.ErrGameTimeoutStart
		if serverExe != "-" {
			_, _ = commonProcess.Kill(serverExe)
		}
		return
	}
	if broadcastBattleServer {
		var port int
		if gameId == common.GameAoE1 {
			port = 8888
		} else {
			port = 9999
		}
		rebroadcastBattleServer(exitCode, port)
	}
	var PID uint32
	for _, p := range processes {
		PID = p
		break
	}
	if waitForProcess(PID) {
		if serverExe != "-" {
			if _, err := commonProcess.Kill(serverExe); err != nil {
				*exitCode = internal.ErrFailedStopServer
			}
		}
	} else {
		*exitCode = internal.ErrFailedWaitForProcess
	}
}
