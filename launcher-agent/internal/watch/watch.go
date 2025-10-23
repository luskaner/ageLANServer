package watch

import (
	"time"

	"github.com/luskaner/ageLANServer/common"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
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

func Watch(gameId string, steamProcess bool, xboxProcess bool, serverExe string, broadcastBattleServer bool,
	battleServerExe string, battleServerRegion string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if serverExe != "-" {
		defer func() {
			if _, err := commonProcess.Kill(serverExe); err != nil {
				if *exitCode == common.ErrSuccess {
					*exitCode = internal.ErrFailedStopServer
				}
			}
			if battleServerExe != "-" && battleServerRegion != "-" {
				newExitCode := launcherCommon.RemoveBattleServerRegion(
					battleServerExe, gameId, battleServerRegion, nil,
				).ExitCode
				if *exitCode == common.ErrSuccess {
					*exitCode = newExitCode
				}
			}
		}()
	}
	defer func() {
		_ = launcherCommon.RunRevertCommand(nil)
	}()
	defer func() {
		launcherCommon.ConfigRevert(gameId, true, nil, nil)
	}()
	processes := waitUntilAnyProcessExist(commonProcess.GameProcesses(gameId, steamProcess, xboxProcess))
	if len(processes) == 0 {
		*exitCode = internal.ErrGameTimeoutStart
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
	if !waitForProcess(PID) {
		*exitCode = internal.ErrFailedWaitForProcess
	}
}
