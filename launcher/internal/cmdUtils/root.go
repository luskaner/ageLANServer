package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherExecutor "github.com/luskaner/ageLANServer/launcher-common/executor"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"runtime"
)

type Config struct {
	gameId          string
	startedAgent    bool
	unmapIPs        bool
	unmapCDN        bool
	removeUserCert  bool
	removeLocalCert bool
	restoreMetadata bool
	restoreProfiles bool
	serverExe       string
	agentStarted    bool
	setupCommandRan bool
	revertCommand   []string
	hostFilePath    string
}

func (c *Config) MappedHosts() {
	c.startedAgent = true
	c.unmapIPs = true
}

func (c *Config) MappedCDN() {
	c.startedAgent = true
	c.unmapCDN = true
}

func (c *Config) LocalCert() {
	c.startedAgent = true
	c.removeLocalCert = true
}

func (c *Config) UserCert() {
	c.removeUserCert = true
}

func (c *Config) BackedUpMetadata() {
	c.restoreMetadata = true
}

func (c *Config) BackedUpProfiles() {
	c.restoreProfiles = true
}

func (c *Config) SetAgentStarted() {
	c.agentStarted = true
}

func (c *Config) SetServerExe(exe string) {
	c.serverExe = exe
}

func (c *Config) SetRevertCommand(cmd []string) {
	c.revertCommand = cmd
}

func (c *Config) SetGameId(id string) {
	c.gameId = id
}

func (c *Config) SetHostFilePath(path string) {
	c.hostFilePath = path
}

func (c *Config) CfgAgentStarted() bool {
	return !commonExecutor.IsAdmin() && c.startedAgent
}

func (c *Config) RequiresConfigRevert() bool {
	return c.unmapIPs || c.unmapCDN || c.removeUserCert || c.removeLocalCert || c.restoreMetadata || c.restoreProfiles
}

func (c *Config) RequiresRunningRevertCommand() bool {
	return c.setupCommandRan && len(c.revertCommand) > 0
}

func (c *Config) AgentStarted() bool {
	return c.agentStarted
}

func (c *Config) ServerExe() string {
	return c.serverExe
}

func (c *Config) RevertCommand() []string {
	if c.setupCommandRan {
		return c.revertCommand
	}
	return []string{}
}

func (c *Config) HostFilePath() string {
	return c.hostFilePath
}

func (c *Config) Revert() {
	if c.AgentStarted() {
		c.KillAgent()
	}
	if serverExe := c.ServerExe(); len(serverExe) > 0 {
		fmt.Println("Stopping 'server'...")
		if proc, err := commonProcess.Kill(serverExe); err == nil {
			fmt.Println("'Server' stopped.")
		} else {
			fmt.Println("Failed to stop 'server'.")
			fmt.Println("Error message: " + err.Error())
			if proc != nil {
				fmt.Println("You may try killing it manually. Kill process 'server' if it is running in your task manager.")
			}
		}
	}
	if c.RequiresConfigRevert() {
		fmt.Println("Cleaning up...")
		if result := executor.RunRevert(c.gameId, c.unmapIPs, c.removeUserCert, c.removeLocalCert, c.restoreMetadata, c.restoreProfiles, c.unmapCDN, c.hostFilePath, true); result.Success() {
			fmt.Println("Cleaned up.")
		} else {
			fmt.Println("Failed to clean up.")
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
		}
	}
	if c.RequiresRunningRevertCommand() {
		err := c.RunRevertCommand()
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
	return
}

func (c *Config) RunRevertCommand() (err error) {
	err = launcherExecutor.RunRevertCommand(c.revertCommand)
	return
}
