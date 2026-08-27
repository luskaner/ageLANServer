package watch

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/gameLogs"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
)

var processWaitInterval = 1 * time.Second
var oneMinuteWaitTimeout = 1 * time.Minute

var (
	waitUntilAnyProcessExistFn = waitUntilAnyProcessExist
	waitForProcessesToExitFn   = waitForProcessesToExit
	serverKillDoFn             = serverKill.Do
	configRevertFn             = launcherCommon.ConfigRevert
	runRevertCommandFn         = launcherCommon.RunRevertCommand
	removeBattleServerRegionFn = launcherCommon.RemoveBattleServerRegion
	gameLogsCopyFn             = gameLogs.CopyGameLogs
	rebroadcastFn              = rebroadcastBattleServer
	commonProcessProcessesByNamesFn = commonProcess.ProcessesByNames
	loggerBufferFn             = func(name string, fn func(io.Writer)) error {
		if internal.Logger == nil {
			return nil
		}
		return internal.Logger.Buffer(name, fn)
	}
)

func waitUntilAnyProcessExist(names []string) (processes map[string]*os.Process) {
	for i := 0; i < int(oneMinuteWaitTimeout/processWaitInterval); i++ {
		processes = commonProcessProcessesByNamesFn(names)
		if len(processes) > 0 {
			return
		}
		time.Sleep(processWaitInterval)
	}
	return
}

func Watch(values *agent.Values, exitCode *int, cleanupOnce *sync.Once) {
	*exitCode = common.ErrSuccess
	defer func() {
		cleanupOnce.Do(func() {
			if values.ServerExecutable != "" {
				commonLogger.Println("Killing server...")
				if err := serverKillDoFn(values.ServerExecutable); err != nil {
					commonLogger.Println("Failed to kill server.")
					commonLogger.Println(err.Error())
					if *exitCode == common.ErrSuccess {
						*exitCode = internal.ErrFailedStopServer
					}
				}
				if values.BattleServerManagerExecutable != "" && values.BattleServerRegion != "" {
					commonLogger.Println("Shutting down battle-server...")
					var result *exec.Result
					if logErr := loggerBufferFn("battle-server-manager_remove", func(writer io.Writer) {
						result = removeBattleServerRegionFn(
							values.BattleServerManagerExecutable, values.GameId, values.BattleServerRegion, writer, func(options *exec.Options) {
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
			}
			_ = loggerBufferFn("revert_command_end", func(writer io.Writer) {
				if err := runRevertCommandFn(writer, func(options *exec.Options) {
					if writer != nil {
						commonLogger.Println("run revert command", options.String())
					}
				}); err != nil {
					commonLogger.Printf("Failed to revert command: %v\n", err)
				}
			})
			_ = loggerBufferFn("config_revert_end", func(writer io.Writer) {
				if !configRevertFn(values.GameId, values.LogRoot, true, writer, func(options *exec.Options) {
					if writer != nil {
						commonLogger.Println("run config revert", options.String())
					}
				}, nil) {
					commonLogger.Println("Failed to revert configuration")
				}
			})
		})
	}()
	commonLogger.Println("Waiting up to 1 minute for game to start...")
	processes := waitUntilAnyProcessExistFn(values.ProcessNames)
	if len(processes) == 0 {
		commonLogger.Println("Failed to find the game.")
		*exitCode = internal.ErrGameTimeoutStart
		return
	}
	if values.BattleServerLANRebroadcast {
		port := battleServer.BroadcastPort(values.GameId)
		commonLogger.Printf("Broadcasting BattleServer port to %d...\n", port)
		rebroadcastFn(exitCode, int(port))
	}
	var procPids []int
	var processesList []*os.Process
	for _, name := range sortedProcessNames(processes) {
		p := processes[name]
		procPids = append(procPids, p.Pid)
		processesList = append(processesList, p)
	}
	commonLogger.Printf("Waiting for PIDs %v to end\n", procPids)
	if !waitForProcessesToExitFn(processesList) {
		commonLogger.Println("Failed to wait.")
		*exitCode = internal.ErrFailedWaitForProcess
		return
	}
	if values.LogRoot != "" && values.BaseDataPath != "" {
		gameLogsCopyFn(values.GameId, values.BaseDataPath, values.LogRoot)
	}
}
