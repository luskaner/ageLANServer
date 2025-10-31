package cmdUtils

import (
	"os"
	"runtime"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/game"
)

func (c *Config) KillAgent() {
	agent := common.GetExeFileName(false, common.LauncherAgent)
	proc, err := commonProcess.Kill(agent)
	if proc != nil {
		logger.Println("Killing 'agent'...")
		if err != nil {
			logger.Println("Failed to kill it: ", err, ", try using the task manager.")
			return
		}
	}
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
		str := "Starting 'agent'"
		if canBroadcastBattleServer == "true" {
			str += ", authorize it in firewall if needed"
		}
		logger.Println(str + "...")
		steamProcess, xboxProcess := executer.GameProcesses()
		var err error
		var f *os.File
		if f, err = commonLogger.FileLogger.Open("agent"); err != nil {
			logger.Println("Error message: " + err.Error())
			return common.ErrFileLog
		}
		loggerPath := commonLogger.FileLogger.Folder()
		if loggerPath == "" {
			loggerPath = "-"
		}
		result := executor.StartAgent(
			c.gameId,
			steamProcess,
			xboxProcess,
			c.serverExe,
			canBroadcastBattleServer == "true",
			c.battleServerExe,
			c.battleServerRegion,
			loggerPath,
			f,
			func(options commonExecutor.Options) {
				commonLogger.Println("start agent", options.String())
			},
		)
		if !result.Success() {
			logger.Println("Failed to start 'agent'.")
			errorCode = internal.ErrAgentStart
			if result.Err != nil {
				logger.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				logger.Printf(`Exit code: %d.`+"\n", result.ExitCode)
			}
			return
		} else {
			logger.Println("'Agent' started.")
		}
	}
	str := "Starting game"
	if customExecutor.Executable != "" {
		str += ", authorize it if needed"
	}
	logger.Println(str + "...")
	var result *commonExecutor.Result
	var values map[string]string = nil
	if c.hostFilePath != "" {
		values = map[string]string{
			"HostFilePath": c.hostFilePath,
		}
		if runtime.GOOS == "windows" {
			values["HostFilePath"] = strings.ReplaceAll(c.hostFilePath, `\`, `\\`)
		}
	}
	if c.certFilePath != "" {
		if values == nil {
			values = make(map[string]string)
		}
		values["CertFilePath"] = c.certFilePath
		if runtime.GOOS == "windows" {
			values["CertFilePath"] = strings.ReplaceAll(c.certFilePath, `\`, `\\`)
		}
	}
	args, err := ParseCommandArgs("Client.ExecutableArgs", values)
	if err != nil {
		logger.Println("Failed to parse client executable arguments")
		errorCode = internal.ErrInvalidClientArgs
		return
	}

	if result = executer.Execute(args, func(options commonExecutor.Options) {
		commonLogger.Println("start game", options.String())
	}); !result.Success() && result.Err != nil {
		if customExecutor.Executable != "" && adminError(result) {
			if canTrustCertificate == "user" {
				logger.Println("Using a user certificate. If it fails to connect to the 'server', try setting the config/option setting 'CanTrustCertificate' to 'local'.")
			}
			result = customExecutor.ExecuteElevated(args, func(options commonExecutor.Options) {
				commonLogger.Println("start elevated game", options.String())
			})
		}
	}
	if !result.Success() {
		errorCode = internal.ErrGameLauncherStart
		if result.Err != nil {
			logger.Println("Game failed to start. Error message: " + result.Err.Error())
		}
		c.KillAgent()
	} else {
		logger.Println("Game started.")
	}
	return
}
