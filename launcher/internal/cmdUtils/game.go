package cmdUtils

import (
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"github.com/luskaner/aoe2DELanServer/launcher/internal"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/executor"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/game"
)

func (c *Config) KillAgent() {
	proc, err := commonProcess.Kill(common.GetExeFileName(false, common.LauncherAgent))
	if err != nil && proc != nil {
		fmt.Println("You may try killing it manually. Search for the process PID inside agent.pid if it exists")
	}
}

func (c *Config) LaunchAgentAndGame(executable string, args []string, canTrustCertificate string, canBroadcastBattleServer string) (errorCode int) {
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
	var broadcastBattleServer bool
	if canBroadcastBattleServer == "auto" && game.RequiresBattleServerBroadcast() {
		canBroadcastBattleServer = "true"
	}
	revertCommand := c.RevertCommand()
	if len(revertCommand) > 0 || broadcastBattleServer || len(c.serverExe) > 0 || c.RequiresConfigRevert() {
		fmt.Print("Starting agent")
		if canBroadcastBattleServer == "true" {
			fmt.Print(", authorize 'agent' in firewall if needed")
		}
		fmt.Println("...")
		steamProcess, microsoftStoreProcess := executer.GameProcesses()
		result := executor.RunAgent(steamProcess, microsoftStoreProcess, c.serverExe, broadcastBattleServer, revertCommand, c.unmapIPs, c.removeUserCert, c.removeLocalCert, c.restoreMetadata, c.restoreProfiles, c.unmapCDN)
		if !result.Success() {
			fmt.Println("Failed to start agent.")
			errorCode = internal.ErrAgentStart
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d. See documentation for "agent" to check what it means.`+"\n", result.ExitCode)
			}
			return
		} else {
			c.SetAgentStarted()
			fmt.Println("Agent started.")
		}
	}
	if _, ok := executer.(game.CustomExecutor); ok {
		fmt.Println("Starting game, authorize it if needed...")
	} else {
		fmt.Println("Starting game...")
	}
	var result *commonExecutor.Result

	if result = executer.Execute(args); !result.Success() && result.Err != nil {
		if customExecutor.Executable != "" && adminError(result) {
			if canTrustCertificate == "user" {
				fmt.Println("Using a user certificate. If it fails to connect to the server, try setting the config/option setting 'CanTrustCertificate' to 'local'.")
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
