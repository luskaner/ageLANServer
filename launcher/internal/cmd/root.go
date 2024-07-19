package cmd

import (
	"common"
	"fmt"
	commonExecutor "launcher-common/executor"
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
	serverPid       uint32
	watcherPid      uint32
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

func (c *Config) SetWatcherPid(pid uint32) {
	c.watcherPid = pid
}

func (c *Config) SetServerPid(pid uint32) {
	c.serverPid = pid
}

func (c *Config) StartedAgent() bool {
	return c.startedAgent
}

func (c *Config) RequiresConfigRevert() bool {
	return c.unmapIPs || c.removeUserCert || c.removeLocalCert || c.restoreMetadata || c.restoreProfiles
}

func (c *Config) WatcherPid() uint32 {
	return c.watcherPid
}

func (c *Config) Revert() {
	if c.serverPid > 0 {
		fmt.Println("Stopping server...")
		if err := commonExecutor.Kill(int(c.serverPid)); err == nil {
			fmt.Println("Server stopped.")
		} else {
			fmt.Println("Failed to stop server.")
			fmt.Println("Error message: " + err.Error())
			fmt.Println("You may need to stop the server manually. Search for the process with the name", common.GetExeFileName(common.Server), "and PID", c.serverPid)
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
