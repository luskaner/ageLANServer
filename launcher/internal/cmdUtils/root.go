package cmdUtils

import (
	"fmt"
	"runtime"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
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
	c.KillAgent()
	if c.serverExe != "" {
		fmt.Println("Stopping 'server'...")
		if proc, err := commonProcess.Kill(c.serverExe); err == nil {
			fmt.Println("'Server' stopped.")
		} else {
			fmt.Println("Failed to stop 'server'.")
			fmt.Println("Error message: " + err.Error())
			if proc != nil {
				fmt.Println("You may try killing it manually. Kill process 'server' if it is running in your task manager.")
			}
		}
	}
	if c.battleServerRegion != "" && c.battleServerExe != "" {
		fmt.Println("Stopping battle server via 'battle-server-manager'...")
		if result := launcherCommon.RemoveBattleServerRegion(c.battleServerExe, c.gameId, c.battleServerRegion); result.Success() {
			fmt.Println("Battle-server stopped (or was already).")
		} else {
			fmt.Println("Failed to stop the battle-server.")
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			fmt.Println("You may try killing it manually. Kill process 'BattleServer.exe' if it is running in your task manager.")
		}
	}
	if c.RequiresConfigRevert() {
		fmt.Println("Cleaning up...")
		if ok := launcherCommon.ConfigRevert(c.gameId, false, executor.RunRevert); !ok {
			fmt.Println("Failed to clean up.")
		}
	}
	if c.RequiresRunningRevertCommand() {
		err := executor.RunRevertCommand()
		if err != nil {
			fmt.Println("Failed to run revert command.")
			fmt.Println("Error message: " + err.Error())
		} else {
			fmt.Println("Ran Revert command.")
		}
	}
}

func anyProcessExists(names []string) bool {
	processes := commonProcess.ProcessesPID(names)
	return len(processes) > 0
}

func GameRunning() bool {
	xbox := runtime.GOOS == "windows"
	for gameId := range common.AllGames.Iter() {
		if anyProcessExists(commonProcess.GameProcesses(gameId, true, xbox)) {
			fmt.Println("Some Age game is already running, exit the game and execute the 'launcher' again.")
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
	result = exec.Options{
		File:           cmd[0],
		Wait:           true,
		SpecialFile:    true,
		Shell:          true,
		UseWorkingPath: true,
		Args:           args,
	}.Exec()
	certStore.ReloadSystemCertificates()
	common.ClearCache()
	return
}
