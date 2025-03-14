package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
)

func (c *Config) KillAgent() {
	proc, err := commonProcess.Kill(common.GetExeFileName(false, common.LauncherAgent))
	fmt.Println(`Stopping 'agent''...`)
	if err != nil && proc != nil {
		fmt.Println("Could not stop it. Kill it using your task manager")
	}
}

func (c *Config) LaunchAgentAndGame(executable string, args []string, canTrustCertificate string, canBroadcastBattleServer string) (errorCode int) {
	fmt.Println("Looking for the game...")
	executer := game.MakeExecutor(c.gameId, executable)
	var customExecutor game.CustomExecutor
	switch executer.(type) {
	case game.SteamExecutor:
		fmt.Println("Game found on Steam.")
	case game.XboxExecutor:
		fmt.Println("Game found on Xbox.")
	case game.CustomExecutor:
		customExecutor = executer.(game.CustomExecutor)
		fmt.Println("Game found on custom path.")
	default:
		fmt.Println("Game not found.")
		errorCode = internal.ErrGameLauncherNotFound
		return
	}
	var broadcastBattleServer bool
	if canBroadcastBattleServer == "auto" && game.RequiresBattleServerBroadcast() {
		canBroadcastBattleServer = "true"
	}
	revertCommand := c.RevertCommand()
	requiresConfigRevert := c.RequiresConfigRevert()
	if len(revertCommand) > 0 || broadcastBattleServer || len(c.serverExe) > 0 || requiresConfigRevert {
		fmt.Print("Starting 'agent'")
		if canBroadcastBattleServer == "true" {
			fmt.Print(", authorize it in firewall if needed")
		}
		fmt.Println("...")
		steamProcess, xboxProcess := executer.GameProcesses()
		var revertFlags []string
		if requiresConfigRevert {
			revertFlags = executor.RevertFlags(c.gameId, c.unmapIPs, c.removeUserCert, c.removeLocalCert, c.restoreMetadata, c.restoreProfiles, c.unmapCDN, true)
		}
		result := executor.RunAgent(c.gameId, steamProcess, xboxProcess, c.serverExe, broadcastBattleServer, revertCommand, revertFlags)
		if !result.Success() {
			fmt.Println("Failed to start 'agent'.")
			errorCode = internal.ErrAgentStart
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			return
		} else {
			c.SetAgentStarted()
			fmt.Println("'Agent' started.")
		}
	}
	fmt.Print("Starting game")
	if _, ok := executer.(game.CustomExecutor); ok {
		fmt.Print(", authorize it if needed")
	}
	fmt.Println("...")
	var result *commonExecutor.Result

	if result = executer.Execute(args); !result.Success() && result.Err != nil {
		if customExecutor.Executable != "" && adminError(result) {
			if canTrustCertificate == "user" {
				fmt.Println("Using a user certificate. If it fails to connect to the 'server', try setting the config/option setting 'CanTrustCertificate' to 'local'.")
			}
			result = customExecutor.ExecuteElevated(args)
		}
	}
	if !result.Success() {
		errorCode = internal.ErrGameLauncherStart
		fmt.Println("Game failed to start. Error message: " + result.Err.Error())
		if c.AgentStarted() {
			c.KillAgent()
		}
	} else {
		fmt.Println("Game started.")
	}
	return
}
