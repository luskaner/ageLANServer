package cmdUtils

import (
	"io"
	"runtime"
	"strconv"
	"time"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/certStore"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonGame "github.com/luskaner/ageLANServer/common/game"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/common/process/game"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/executor"
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
)

type Config struct {
	// Report is where the run says what it is doing.
	//
	// It is a field rather than a package global so that two things can hold
	// this type at once without talking over each other: a window running a
	// session while a test runs another, or two windows side by side. A frontend
	// installs one before the run starts; everything below reports through it,
	// which is what lets the same run be read in a terminal and in a window.
	Report launcher.Reporter

	gameId             string
	serverExe          string
	setupCommandRan    bool
	hostFilePath       string
	certFilePath       string
	battleServerRegion string
	battleServerExe    string
}

// report is the Reporter this config was given, or one that discards.
//
// The zero Config has no Reporter, and it is a legitimate value: tests build one
// to exercise a single step, and a caller that does not care about output should
// not have to invent a Reporter to get one. Falling back here rather than
// panicking keeps that ordinary.
func (c *Config) report() launcher.Reporter {
	if c.Report == nil {
		return launcher.Discard{}
	}
	return c.Report
}
func (c *Config) SetGameId(gameId string) {
	c.gameId = gameId
}

func (c *Config) RequiresConfigRevert() bool {
	if err, args := launcherCommon.RevertConfigStore.Load(); err == nil && len(args) > 0 {
		return true
	}
	return false
}

func (c *Config) revertCommand() []string {
	if err, args := launcherCommon.RevertCommandStore.Load(); err == nil {
		return args
	}
	return []string{}
}

func (c *Config) RequiresRunningRevertCommand() bool {
	return c.setupCommandRan && len(c.revertCommand()) > 0
}

// HasTeardownWork reports whether Revert has anything it would show the user.
//
// It exists so the caller can decide whether a phase heading is honest. Revert
// runs on every exit, but on a clean run each step already undid its own changes,
// so all that is left is a kill that finds nothing, a log close and a lock
// release: three things that print nothing. Announcing a phase for that is worse
// than staying quiet, because the reader is told to wait for work that does not
// exist.
func (c *Config) HasTeardownWork() bool {
	return c.AgentRunning() ||
		c.serverExe != "" ||
		(c.battleServerRegion != "" && c.battleServerExe != "") ||
		c.RequiresConfigRevert() ||
		c.RequiresRunningRevertCommand()
}

func (c *Config) RevertCommand() []string {
	if c.setupCommandRan {
		return c.revertCommand()
	}
	return []string{}
}

func (c *Config) Revert() {
	logger.WriteFileLog(c.gameId, "pre-revert")
	c.KillAgent()
	if c.serverExe != "" {
		c.report().Step("Stopping server...")
		if err := serverKill.Do(c.serverExe); err == nil {
			c.report().Ok("Server stopped.")
		} else {
			c.report().Fail("Failed to stop server.")
			c.report().Fault("Error message: %s", err.Error())
		}
	}
	if c.battleServerRegion != "" && c.battleServerExe != "" {
		c.report().Step("Stopping battle server via battle-server-manager...")
		_ = commonLogger.FileLogger.Buffer("battle-server-manager_remove", func(writer io.Writer) {
			if result := launcherCommon.RemoveBattleServerRegion(c.battleServerExe, c.gameId, c.battleServerRegion, writer, func(options *exec.Options) {
				commonLogger.Println("battle-server-manager_remove", options.String())
			}); result.Success() {
				c.report().Ok("Battle-server stopped (or was already).")
			} else {
				c.report().Fail("Failed to stop the battle-server.")
				if result.Err != nil {
					c.report().Fault("Error message: %s", result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					c.report().Fault("Exit code: %d.", result.ExitCode)
				}
				c.report().Fault("You may try killing it manually. Kill process %s if it is running in your task manager.", battleServer.Executable)
			}
		})
	}
	if c.RequiresConfigRevert() {
		c.report().Step("Cleaning up...")
		_ = commonLogger.FileLogger.Buffer("config_revert", func(writer io.Writer) {
			if ok := launcherCommon.ConfigRevert(c.gameId, commonLogger.FileLogger.Folder(), false, writer, func(options *exec.Options) {
				commonLogger.Println("run config revert", options.String())
			}, executor.RunRevert); !ok {
				c.report().Fail("Failed to clean up.")
			}
		})
	} else if launcherCommon.ConfigAdminAgentRunning(false) {
		c.report().Step("Stopping config-admin-agent...")
		if result := c.RunStopAgent(); result.Success() {
			c.report().Ok("Config-admin-agent stopped.")
		} else {
			c.report().Fail("Failed to stop agent.")
			if result.Err != nil {
				c.report().Fault("Error message: %s", result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				c.report().Fault("Exit code: %s", strconv.Itoa(result.ExitCode))
			}
		}
	}
	if c.RequiresRunningRevertCommand() {
		_ = commonLogger.FileLogger.Buffer("revert_command", func(writer io.Writer) {
			err := executor.RunRevertCommand(writer, func(options *exec.Options) {
				commonLogger.Println("run revert command", options.String())
			})
			if err != nil {
				c.report().Fail("Failed to run revert command.")
				c.report().Fault("Error message: %s", err.Error())
			} else {
				c.report().Ok("Ran Revert command.")
			}
		})
	}
	logger.WriteFileLog(c.gameId, "post-revert")
}

func anyProcessExists(names []string) bool {
	processes := commonProcess.ProcessesByNames(names)
	return len(processes) > 0
}

func GameRunning(r launcher.Reporter) bool {
	xbox := runtime.GOOS == "windows"
	steamMacOsNative := runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
	var gameProcesses []string
	for gameId := range commonGame.AllGames.Iter() {
		gameProcesses = append(gameProcesses, game.Processes(gameId, true, steamMacOsNative, xbox)...)
	}
	someProcessRunning := func() bool {
		return anyProcessExists(gameProcesses)
	}
	if !someProcessRunning() {
		return false
	}
	r.Step("Some Age game is already running, waiting up to 1 minute for the game to exit.")
	timeout := time.After(1 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			r.Warn("The game did not exit in time.")
			return true
		case <-ticker.C:
			if !someProcessRunning() {
				return false
			}
		}
	}
}

func (c *Config) RunSetupCommand(cmd []string) (result *exec.Result) {
	var args []string
	if len(cmd) > 1 {
		args = cmd[1:]
	}
	options := exec.Options{
		File:           cmd[0],
		Wait:           true,
		SpecialFile:    true,
		UseWorkingPath: true,
		Args:           args,
		ExitCode:       true,
	}
	if buffErr := commonLogger.FileLogger.Buffer("setup_command", func(writer io.Writer) {
		commonLogger.Println("run setup command", options.String())
		if writer != nil {
			options.Stderr = writer
			options.Stdout = writer
		}
		result = options.Exec()
	}); buffErr != nil {
		result.Err = buffErr
		result.ExitCode = common.ErrFileLog
	}
	certStore.ReloadSystemCertificates()
	common.ClearDNSCache()
	return
}
