package cmd

import (
	"common"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows"
	commonExecutor "launcher-common/executor"
	"launcher/internal"
	"launcher/internal/executor"
	"launcher/internal/game"
)

func (c *Config) KillWatcher() {
	err := commonExecutor.Kill(int(c.watcherPid))
	if err != nil {
		fmt.Printf(`Failed to terminate watcher. Kill it manually if it still exists, search for name "%s" and/or PID %d\n`, common.GetExeFileName(true, common.LauncherWatcher), c.watcherPid)
	}
}

func (c *Config) LaunchWatcherAndGame(executable string, canTrustCertificate string) (errorCode int) {
	fmt.Println("Looking for the game...")
	executer := game.MakeExecutor(executable)
	var customExecutor game.CustomExecutor
	switch executer.(type) {
	case game.SteamExecutor:
		fmt.Println("Game found on Steam.")
	case game.MicrosoftStoreExecutor:
		fmt.Println("Game found on Microsoft Store.")
	case game.CustomExecutor:
		customExecutor = executer.(game.CustomExecutor)
		fmt.Println("Game found on custom path.")
	default:
		fmt.Println("Game not found.")
		errorCode = internal.ErrGameLauncherNotFound
		return
	}
	if c.serverPid > 0 || c.RequiresConfigRevert() {
		fmt.Println("Starting watcher...")
		result := executor.RunWatcher(executer.FinalExecutable(), c.serverPid, c.unmapIPs, c.removeUserCert, c.removeLocalCert, c.restoreMetadata, c.restoreProfiles)
		if !result.Success() {
			fmt.Println("Failed to start watcher.")
			errorCode = internal.ErrWatcherStart
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d. See documentation for "watcher" to check what it means.`+"\n", result.ExitCode)
			}
			return
		} else {
			c.SetWatcherPid(result.Pid)
			fmt.Println("Watcher started.")
		}
	}

	fmt.Println("Starting game...")
	var result *commonExecutor.ExecResult
	executableArgs := viper.GetStringSlice("Client.ExecutableArgs")

	if result = executer.Execute(executableArgs); !result.Success() && result.Err != nil {
		if customExecutor.Executable != "" && errors.Is(result.Err, windows.ERROR_ELEVATION_REQUIRED) {
			if canTrustCertificate == "user" {
				fmt.Println("Using a user certificate. If it fails to connect to the server, try setting the config/option setting 'CanTrustCertificate' to 'local'.")
			}
			result = customExecutor.ExecuteElevated(executableArgs)
		}
	}
	if !result.Success() {
		errorCode = internal.ErrGameLauncherStart
		fmt.Println("Game failed to start. Error message: " + result.Err.Error())
		if c.watcherPid > 0 {
			c.KillWatcher()
		}
	} else {
		fmt.Println("Game started.")
	}
	return
}
