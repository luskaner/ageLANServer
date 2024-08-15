package cmdUtils

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	launcherExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/executor"
)

type Config struct {
	startedAgent    bool
	unmapIPs        bool
	removeUserCert  bool
	removeLocalCert bool
	restoreMetadata bool
	restoreProfiles bool
	serverExe       string
	agentStarted    bool
	setupCommandRan bool
	revertCommand   []string
}

func (c *Config) MappedHosts() {
	c.startedAgent = true
	c.unmapIPs = true
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

func (c *Config) CfgAgentStarted() bool {
	return c.startedAgent
}

func (c *Config) RequiresConfigRevert() bool {
	return c.unmapIPs || c.removeUserCert || c.removeLocalCert || c.restoreMetadata || c.restoreProfiles
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

func (c *Config) Revert() {
	if c.AgentStarted() {
		c.KillAgent()
	}
	if serverExe := c.ServerExe(); len(serverExe) > 0 {
		fmt.Println("Stopping server...")
		if proc, err := commonProcess.Kill(serverExe); err == nil {
			fmt.Println("Server stopped.")
		} else {
			fmt.Println("Failed to stop server.")
			fmt.Println("Error message: " + err.Error())
			if proc != nil {
				fmt.Println("You may try killing it manually. Search for the process PID inside server.pid if it exists")
			}
		}
	}
	if c.RequiresConfigRevert() {
		fmt.Println("Cleaning up...")
		if result := executor.RunRevert(c.unmapIPs, c.removeUserCert, c.removeLocalCert, c.restoreMetadata, c.restoreProfiles); result.Success() {
			fmt.Println("Cleaned up.")
		} else {
			fmt.Println("Failed to clean up.")
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d. See documentation for "config" to check what it means`+"\n", result.ExitCode)
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

func GameRunning() bool {
	if commonProcess.AnyProcessExists(commonProcess.GameProcesses(true, true)) {
		fmt.Println("Game is already running, exiting...")
		return true
	}
	return false
}

func (c *Config) RunSetupCommand(cmd []string) (err error) {
	err = launcherExecutor.RunCommand(cmd)
	if err == nil {
		c.setupCommandRan = true
	}
	return
}

func (c *Config) RunRevertCommand() (err error) {
	err = launcherExecutor.RunCommand(c.revertCommand)
	return
}
