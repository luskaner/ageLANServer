package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
	"runtime"
	"strings"
)

func (c *Config) KillAgent() bool {
	agent := common.GetExeFileName(false, common.LauncherAgent)
	proc, err := commonProcess.Kill(agent)
	if proc != nil {
		fmt.Println("Killing 'agent'...")
		if err != nil {
			fmt.Println("Failed to kill it: ", err, ", try using the task manager.")
			return false
		}
	}
	return true
}

func (c *Config) LaunchAgentAndGame(executer game.Executor, customExecutor game.CustomExecutor, canTrustCertificate string, canBroadcastBattleServer string) (errorCode int) {
	if canBroadcastBattleServer != "false" {
		if game.RequiresBattleServerBroadcast() {
			canBroadcastBattleServer = "true"
		} else {
			canBroadcastBattleServer = "false"
		}
	}
	revertCommand := c.RevertCommand()
	requiresConfigRevert := c.RequiresConfigRevert()
	if len(revertCommand) > 0 || canBroadcastBattleServer == "true" || len(c.serverExe) > 0 || requiresConfigRevert {
		fmt.Print("Starting 'agent'")
		if canBroadcastBattleServer == "true" {
			fmt.Print(", authorize it in firewall if needed")
		}
		fmt.Println("...")
		steamProcess, xboxProcess := executer.GameProcesses()
		result := executor.RunAgent(c.gameId, steamProcess, xboxProcess, c.serverExe, canBroadcastBattleServer == "true")
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
			fmt.Println("'Agent' started.")
		}
	}
	fmt.Print("Starting game")
	if customExecutor.Executable != "" {
		fmt.Print(", authorize it if needed")
	}
	fmt.Println("...")
	var result *commonExecutor.Result
	var values map[string]string = nil
	if hostFilePath := c.HostFilePath(); hostFilePath != "" {
		values = map[string]string{
			"HostFilePath": hostFilePath,
		}
		if runtime.GOOS == "windows" {
			values["HostFilePath"] = strings.ReplaceAll(hostFilePath, `\`, `\\`)
		}
	}
	if certFilePath := c.CertFilePath(); certFilePath != "" {
		if values == nil {
			values = make(map[string]string)
		}
		values["CertFilePath"] = certFilePath
		if runtime.GOOS == "windows" {
			values["CertFilePath"] = strings.ReplaceAll(certFilePath, `\`, `\\`)
		}
	}
	args, err := ParseCommandArgs("Client.ExecutableArgs", values)
	if err != nil {
		fmt.Println("Failed to parse client executable arguments")
		errorCode = internal.ErrInvalidClientArgs
		return
	}

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
		if result.Err != nil {
			fmt.Println("Game failed to start. Error message: " + result.Err.Error())
		}
		c.KillAgent()
	} else {
		fmt.Println("Game started.")
	}
	return
}
