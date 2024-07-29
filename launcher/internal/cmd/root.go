package cmd

import (
	"common"
	commonProcess "common/process"
	"fmt"
	"launcher/internal/executor"
	"launcher/internal/game"
)

type Config struct {
	startedAgent    bool
	unmapIPs        bool
	removeUserCert  bool
	removeLocalCert bool
	restoreMetadata bool
	restoreProfiles bool
	serverExe       string
	watcherStarted  bool
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

func (c *Config) SetWatcherStarted() {
	c.watcherStarted = true
}

func (c *Config) SetServerExe(exe string) {
	c.serverExe = exe
}

func (c *Config) AgentStarted() bool {
	return c.startedAgent
}

func (c *Config) RequiresConfigRevert() bool {
	return c.unmapIPs || c.removeUserCert || c.removeLocalCert || c.restoreMetadata || c.restoreProfiles
}

func (c *Config) WatcherStarted() bool {
	return c.watcherStarted
}

func (c *Config) ServerExe() string {
	return c.serverExe
}

func (c *Config) Revert() {
	if serverExe := c.ServerExe(); len(serverExe) > 0 {
		fmt.Println("Stopping server...")
		if proc, err := commonProcess.Kill(serverExe); err == nil {
			fmt.Println("Server stopped.")
		} else {
			fmt.Println("Failed to stop server.")
			fmt.Println("Error message: " + err.Error())
			if proc != nil {
				fmt.Println("You may try killing it manually. Search for the process with PID", proc.Pid)
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
}

func GameRunning() bool {
	if game.AnyProcessExists(true, true) {
		fmt.Println("Game is already running, exiting...")
		return true
	}
	return false
}
