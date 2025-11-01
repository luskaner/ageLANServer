package cmdUtils

import (
	"io"
	"runtime"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/serverKill"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
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
				logger.Println("You may try killing it manually. Kill process 'BattleServer.exe' if it is running in your task manager.")
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
	processes := commonProcess.ProcessesPID(names)
	return len(processes) > 0
}

func GameRunning() bool {
	xbox := runtime.GOOS == "windows"
	for gameId := range common.AllGames.Iter() {
		if anyProcessExists(commonProcess.GameProcesses(gameId, true, xbox)) {
			logger.Println("Some Age game is already running, exit the game and execute the 'launcher' again.")
			return true
		}
	}
	return false
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
		Shell:          true,
		UseWorkingPath: true,
		Args:           args,
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
	common.ClearCache()
	return
}
