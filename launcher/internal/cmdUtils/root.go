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
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

type Config struct {
	gameId             string
	serverExe          string
	setupCommandRan    bool
	hostFilePath       string
	certFilePath       string
	battleServerRegion string
	battleServerExe    string
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
		logger.Println("Stopping 'server'...")
		if err := serverKill.Do(c.serverExe); err == nil {
			logger.Println("'Server' stopped.")
		} else {
			logger.Println("Failed to stop 'server'.")
			logger.Println("Error message: " + err.Error())
		}
	}
	if c.battleServerRegion != "" && c.battleServerExe != "" {
		logger.Println("Stopping battle server via 'battle-server-manager'...")
		_ = commonLogger.FileLogger.Buffer("battle-server-manager_remove", func(writer io.Writer) {
			if result := launcherCommon.RemoveBattleServerRegion(c.battleServerExe, c.gameId, c.battleServerRegion, writer, func(options exec.Options) {
				commonLogger.Println("battle-server-manager_remove", options.String())
			}); result.Success() {
				logger.Println("Battle-server stopped (or was already).")
			} else {
				logger.Println("Failed to stop the battle-server.")
				if result.Err != nil {
					logger.Println("Error message: " + result.Err.Error())
				}
				if result.ExitCode != common.ErrSuccess {
					logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
				}
				logger.Printf("You may try killing it manually. Kill process '%s' if it is running in your task manager.\n", battleServer.Executable)
			}
		})
	}
	if c.RequiresConfigRevert() {
		logger.Println("Cleaning up...")
		_ = commonLogger.FileLogger.Buffer("config_revert", func(writer io.Writer) {
			if ok := launcherCommon.ConfigRevert(c.gameId, commonLogger.FileLogger.Folder(), false, writer, func(options exec.Options) {
				commonLogger.Println("run config revert", options.String())
			}, executor.RunRevert); !ok {
				logger.Println("Failed to clean up.")
			}
		})
	} else if launcherCommon.ConfigAdminAgentRunning(false) {
		logger.Println("Stopping 'config-admin-agent'...")
		if result := c.RunStopAgent(); result.Success() {
			logger.Println("'Config-admin-agent' stopped.")
		} else {
			logger.Println("Failed to stop agent.")
			if result.Err != nil {
				logger.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Println("Exit code: " + strconv.Itoa(result.ExitCode))
			}
		}
	}
	if c.RequiresRunningRevertCommand() {
		_ = commonLogger.FileLogger.Buffer("revert_command", func(writer io.Writer) {
			err := executor.RunRevertCommand(writer, func(options exec.Options) {
				commonLogger.Println("run revert command", options.String())
			})
			if err != nil {
				logger.Println("Failed to run revert command.")
				logger.Println("Error message: " + err.Error())
			} else {
				logger.Println("Ran Revert command.")
			}
		})
	}
	logger.WriteFileLog(c.gameId, "post-revert")
}

func anyProcessExists(names []string) bool {
	processes := commonProcess.ProcessesByNames(names)
	return len(processes) > 0
}

func GameRunning() bool {
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
	logger.Println("Some Age game is already running, waiting up to 1 minute for the game to exit.")
	timeout := time.After(1 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			logger.Println("The game did not exit in time.")
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
