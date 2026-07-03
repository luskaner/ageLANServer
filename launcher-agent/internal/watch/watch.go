package watch

import (
	"io"
	"os"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/gameLogs"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
)

var processWaitInterval = 1 * time.Second

func waitUntilAnyProcessExist(names []string) (processes map[string]*os.Process) {
	for i := 0; i < int((1*time.Minute)/processWaitInterval); i++ {
		processes = commonProcess.ProcessesByNames(names)
		if len(processes) > 0 {
			return
		}
		time.Sleep(processWaitInterval)
	}
	return
}

func Watch(values *agent.Values, exitCode *int) {
	*exitCode = common.ErrSuccess
	if values.ServerExecutable != "" {
		defer func() {
			commonLogger.Println("Killing server...")
			if err := serverKill.Do(values.ServerExecutable); err != nil {
				commonLogger.Println("Failed to kill server.")
				commonLogger.Println(err.Error())
				if *exitCode == common.ErrSuccess {
					*exitCode = internal.ErrFailedStopServer
				}
			}
			if values.BattleServerManagerExecutable != "" && values.BattleServerRegion != "" {
				commonLogger.Println("Shutting down battle-server...")
				var result *exec.Result
				if logErr := internal.Logger.Buffer("battle-server-manager_remove", func(writer io.Writer) {
					result = launcherCommon.RemoveBattleServerRegion(
						values.BattleServerManagerExecutable, values.GameId, values.BattleServerRegion, writer, func(options exec.Options) {
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
			if !launcherCommon.ConfigRevert(values.GameId, values.LogRoot, true, writer, func(options exec.Options) {
				if writer != nil {
					commonLogger.Println("run config revert", options.String())
				}
			}, nil) {
				commonLogger.Println("Failed to revert configuration")
			}
		})
	}()
	commonLogger.Println("Waiting up to 1 minute for game to start...")
	processes := waitUntilAnyProcessExist(values.ProcessNames)
	if len(processes) == 0 {
		commonLogger.Println("Failed to find the game.")
		*exitCode = internal.ErrGameTimeoutStart
		return
	}
	if values.BattleServerLANRebroadcast {
		port := battleServer.BroadcastPort(values.GameId)
		commonLogger.Printf("Broadcasting BattleServer port to %d...\n", port)
		rebroadcastBattleServer(exitCode, int(port))
	}
	var proc *os.Process
	for _, p := range processes {
		proc = p
		break
	}
	//goland:noinspection ALL
	commonLogger.Printf("Waiting for PID %d to end\n", proc.Pid)
	if !commonProcess.WaitForProcess(proc, nil) {
		commonLogger.Println("Failed to wait.")
		*exitCode = internal.ErrFailedWaitForProcess
		return
	}
	if values.LogRoot != "" && values.BaseDataPath != "" {
		gameLogs.CopyGameLogs(values.GameId, values.BaseDataPath, values.LogRoot)
	}
}
