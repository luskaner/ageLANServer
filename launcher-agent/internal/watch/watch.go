package watch

import (
	"io"
	"os"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/logger"
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

func Watch(gameId string, logRoot string, steamProcess bool, xboxProcess bool, serverExe string, broadcastBattleServer bool,
	battleServerExe string, battleServerRegion string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if serverExe != "-" {
		defer func() {
			commonLogger.Println("Killing server...")
			if _, err := commonProcess.Kill(serverExe); err != nil {
				commonLogger.Println("Failed to kill server.")
				commonLogger.Println(err.Error())
				if *exitCode == common.ErrSuccess {
					*exitCode = internal.ErrFailedStopServer
				}
			}
			if battleServerExe != "-" && battleServerRegion != "-" {
				commonLogger.Println("Shutting down battle-server...")
				var result *exec.Result
				if logErr := internal.Logger.Buffer("battle-server-manager_remove", func(writer io.Writer) {
					result = launcherCommon.RemoveBattleServerRegion(
						battleServerExe, gameId, battleServerRegion, writer, func(options exec.Options) {
							if writer != nil {
								commonLogger.Println("run battle-server-manager", options.String())
							}
						},
					)
				}); logErr != nil {
					result.ExitCode = common.ErrFileLog
					result.Err = logErr
				}
				newExitCode := result.ExitCode
				if !result.Success() {
					commonLogger.Println("Failed to shut down battle-server.")
					if result.ExitCode != common.ErrSuccess {
						commonLogger.Println("Exit code: ", newExitCode)
					}
					if result.Err != nil {
						commonLogger.Printf("Error: %v\n", result.Err)
					}
				}
				if *exitCode == common.ErrSuccess {
					*exitCode = newExitCode
				}
			}
		}()
	}
	defer func() {
		_ = internal.Logger.Buffer("revert_command_end", func(writer io.Writer) {
			if err := launcherCommon.RunRevertCommand(writer, func(options exec.Options) {
				if writer != nil {
					commonLogger.Println("run revert command", options.String())
				}
			}); err != nil {
				commonLogger.Printf("Failed to revert command: %v\n", err)
			}
		})
	}()
	defer func() {
		_ = internal.Logger.Buffer("config_revert_end", func(writer io.Writer) {
			if !launcherCommon.ConfigRevert(gameId, logRoot, true, writer, func(options exec.Options) {
				if writer != nil {
					commonLogger.Println("run config revert", options.String())
				}
			}, nil) {
				commonLogger.Println("Failed to revert configuration")
			}
		})
	}()
	commonLogger.Println("Waiting up to 1 minute for game to start...")
	processes := waitUntilAnyProcessExist(commonProcess.GameProcesses(gameId, steamProcess, xboxProcess))
	if len(processes) == 0 {
		commonLogger.Println("Failed to find the game.")
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
		commonLogger.Printf("Broadcasting BattleServer port to %d...\n", port)
		rebroadcastBattleServer(exitCode, port)
	}
	var PID uint32
	for _, p := range processes {
		PID = p
		break
	}
	commonLogger.Printf("Waiting for PID %d to end\n", PID)
	if !commonProcess.WaitForProcess(&os.Process{Pid: int(PID)}, nil) {
		commonLogger.Println("Failed to wait.")
		*exitCode = internal.ErrFailedWaitForProcess
	}
}
