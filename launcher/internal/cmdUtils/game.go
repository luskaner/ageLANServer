package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/config/launcher/parse"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	commonExecutor "github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
	"net"
	"runtime"
	"strings"
)

func (c *Config) KillAgent() bool {
	agent := common.GetExeFileName(false, common.LauncherAgent)
	proc, err := commonProcess.Kill(agent)
	if proc != nil {
		fmt.Print(
			printer.Gen(
				printer.Stop,
				"",
				printer.T("Stopping "),
				printer.TS("agent", printer.ComponentStyle),
				printer.T("... "),
			),
		)
		if err != nil {
			printer.PrintFailedError(err)
			return false
		} else {
			printer.PrintSucceeded()
		}
	}
	return true
}

func (c *Config) ParseGameArguments(rawArgs []string) (args []string) {
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
	args = parse.CommandArgs(rawArgs, values)
	return
}

func (c *Config) LaunchAgent(executer game.Executor, rebroadcastIPs []net.IP) (errorCode int) {
	revertCommand := c.RevertCommand()
	requiresConfigRevert := c.RequiresConfigRevert()
	if len(revertCommand) > 0 || len(rebroadcastIPs) > 0 || c.serverPid != 0 || requiresConfigRevert {
		agentStyledTexts := []*printer.StyledText{
			printer.T("Starting "),
			printer.TS("agent", printer.ComponentStyle),
		}
		if len(rebroadcastIPs) > 0 {
			agentStyledTexts = append(agentStyledTexts, printer.T(", authorize it in firewall if needed"))
		}
		agentStyledTexts = append(agentStyledTexts, printer.T("... "))
		fmt.Print(printer.Gen(printer.Execute, "", agentStyledTexts...))
		steamProcess, xboxProcess := executer.GameProcesses()
		result := executor.RunAgent(c.gameTitle, steamProcess, xboxProcess, c.serverPid, rebroadcastIPs)
		if !result.Success() {
			printer.PrintFailedResultError(result)
			return internal.ErrAgentStart
		} else {
			printer.PrintSucceeded()
		}
	}
	return
}

func (c *Config) LaunchGame(executer game.Executor, customExecutor game.CustomExecutor, clientExecutableArgs []string) (errorCode int) {
	gameStyledTexts := []*printer.StyledText{
		printer.T("Starting game title"),
	}
	if customExecutor.Executable != "" {
		gameStyledTexts = append(gameStyledTexts, printer.T(" , authorize it if needed"))
	}
	gameStyledTexts = append(gameStyledTexts, printer.T("... "))
	fmt.Print(printer.Gen(printer.Execute, "", gameStyledTexts...))
	var result *commonExecutor.Result
	if result = executer.Execute(clientExecutableArgs); !result.Success() && result.Err != nil {
		if customExecutor.Executable != "" && adminError(result) {
			result = customExecutor.ExecuteElevated(clientExecutableArgs)
		}
	}
	if !result.Success() {
		errorCode = internal.ErrGameLauncherStart
		printer.PrintResultError(result)
		c.KillAgent()
	} else {
		printer.PrintSucceeded()
	}
	return
}
