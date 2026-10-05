package watch

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"github.com/luskaner/ageLANServer/launcher-agent/internal/gameLogs"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/agent"
	"github.com/luskaner/ageLANServer/launcher-common/cmdlog"
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
)

var processWaitInterval = 1 * time.Second
var oneMinuteWaitTimeout = 1 * time.Minute

var (
	waitUntilAnyProcessExistFn      = waitUntilAnyProcessExist
	waitForProcessesToExitFn        = waitForProcessesToExit
	serverKillDoFn                  = serverKill.Do
	configRevertFn                  = launcherCommon.ConfigRevert
	runRevertCommandFn              = launcherCommon.RunRevertCommand
	removeBattleServerRegionFn      = launcherCommon.RemoveBattleServerRegion
	gameLogsCopyFn                  = gameLogs.CopyGameLogs
	rebroadcastFn                   = rebroadcastBattleServer
	commonProcessProcessesByNamesFn = commonProcess.ProcessesByNames
	loggerBufferFn                  = func(name string, fn func(io.Writer)) error {
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

// ExitCode carries the agent's exit code between the goroutine that waits for
// the game to exit and the signal handler that can interrupt it, plus the
// cleanup whichever of them gets to run first.
//
// It used to be a bare *int written from both goroutines, which is a data race.
// It went unnoticed on Windows because the agent is started detached with
// CREATE_NO_WINDOW, so it has no console, Ctrl+C never reaches it and SIGTERM
// does not exist: the second writer is only reachable on the platforms where
// the signal actually fires.
type ExitCode struct {
	code atomic.Int32
}

func NewExitCode() *ExitCode {
	e := &ExitCode{}
	e.code.Store(int32(common.ErrSuccess))
	return e
}

func (e *ExitCode) Get() int {
	return int(e.code.Load())
}

// Set records code unconditionally. Reserved for the signal handler: an
// interrupt arrives out of band and has to be recorded even if something was
// already recorded.
func (e *ExitCode) Set(code int) {
	e.code.Store(int32(code))
}

// SetIfSuccess records code only while nothing has failed yet, so the first
// failure is the one that survives: a later step reporting a different error
// must not mask an earlier one. This is what every failure discovered while
// watching or cleaning up uses, including the signal handler's neighbours, so
// that an interrupt recorded first is not overwritten by the fallout it
// causes.
func (e *ExitCode) SetIfSuccess(code int) {
	e.code.CompareAndSwap(int32(common.ErrSuccess), int32(code))
}

// Cleanup is the agent's teardown: stop the server, then the battle server, the
// revert command and finally the configuration. The order matters, stopping the
// server before pulling the configuration leaves the game talking to a live
// server whose certificate trust is being removed underneath it.
//
// This used to exist twice with different orders (here and in cmd.runRoot's
// signal handler), so which one ran depended on how the agent was asked to stop.
func Cleanup(values *agent.Values, exitCode *ExitCode) {
	if values.ServerExecutable != "" {
		cmdlog.Step("Killing server...")
		if err := serverKillDoFn(values.ServerExecutable); err != nil {
			cmdlog.Fail("Failed to kill server.")
			cmdlog.Fault("%s", err.Error())
			exitCode.SetIfSuccess(internal.ErrFailedStopServer)
		}
		if values.BattleServerManagerExecutable != "" && values.BattleServerRegion != "" {
			cmdlog.Step("Shutting down battle-server...")
			var result *exec.Result
			logErr := loggerBufferFn("battle-server-manager_remove", func(writer io.Writer) {
				result = removeBattleServerRegionFn(
					values.BattleServerManagerExecutable, values.GameId, values.BattleServerRegion, writer, func(options *exec.Options) {
						if writer != nil {
							cmdlog.Println("run battle-server-manager", options.String())
						}
					},
				)
			})
			// Guard against a nil result before dereferencing it: the previous
			// inline version panicked here when the log buffer or the call
			// itself failed.
			if result == nil {
				result = &exec.Result{}
			}
			if logErr != nil {
				result.ExitCode = common.ErrFileLog
				result.Err = logErr
			}
			newExitCode := result.ExitCode
			if !result.Success() {
				cmdlog.Fail("Failed to shut down battle-server.")
				if result.ExitCode != common.ErrSuccess {
					cmdlog.Fault("Exit code: %d", newExitCode)
				}
				if result.Err != nil {
					cmdlog.Fault("Error: %s", result.Err)
				}
			}
			exitCode.SetIfSuccess(newExitCode)
		}
	}
	_ = loggerBufferFn("revert_command_end", func(writer io.Writer) {
		if err := runRevertCommandFn(writer, func(options *exec.Options) {
			cmdlog.Println("run revert command", options.String())
		}); err != nil {
			cmdlog.Fail("Failed to revert command: %s", err)
		}
	})
	_ = loggerBufferFn("config_revert_end", func(writer io.Writer) {
		if !configRevertFn(values.GameId, values.LogRoot, true, writer, func(options *exec.Options) {
			if writer != nil {
				cmdlog.Println("run config revert", options.String())
			}
		}, nil) {
			cmdlog.Fail("Failed to revert configuration")
		}
	})
}

func Watch(values *agent.Values, exitCode *ExitCode, cleanupOnce *sync.Once) {
	defer func() {
		cleanupOnce.Do(func() {
			Cleanup(values, exitCode)
		})
	}()
	cmdlog.Step("Waiting up to 1 minute for game to start...")
	processes := waitUntilAnyProcessExistFn(values.ProcessNames)
	if len(processes) == 0 {
		cmdlog.Fail("Failed to find the game.")
		exitCode.SetIfSuccess(internal.ErrGameTimeoutStart)
		return
	}
	if values.BattleServerLANRebroadcast {
		port := battleServer.BroadcastPort(values.GameId)
		cmdlog.Step("Broadcasting BattleServer port to %d...", port)
		rebroadcastFn(exitCode, int(port))
	}
	var procPids []int
	var processesList []*os.Process
	for _, name := range sortedProcessNames(processes) {
		p := processes[name]
		procPids = append(procPids, p.Pid)
		processesList = append(processesList, p)
	}
	cmdlog.Step("Waiting for PIDs %v to end", procPids)
	if !waitForProcessesToExitFn(processesList) {
		cmdlog.Fail("Failed to wait.")
		exitCode.SetIfSuccess(internal.ErrFailedWaitForProcess)
		return
	}
	if values.LogRoot != "" && values.BaseDataPath != "" {
		gameLogsCopyFn(values.GameId, values.BaseDataPath, values.LogRoot)
	}
}
