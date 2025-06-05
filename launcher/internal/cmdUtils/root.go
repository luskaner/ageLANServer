package cmdUtils

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/printer"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
	"github.com/luskaner/ageLANServer/launcher/internal/server/certStore"
	"runtime"
)

type Config struct {
	gameId          string
	serverExe       string
	setupCommandRan bool
	hostFilePath    string
	certFilePath    string
}

func (c *Config) SetServerExe(exe string) {
	c.serverExe = exe
}

func (c *Config) SetGameId(id string) {
	c.gameId = id
}

func (c *Config) SetHostFilePath(path string) {
	c.hostFilePath = path
}

func (c *Config) SetCertFilePath(path string) {
	c.certFilePath = path
}

func (c *Config) RequiresConfigRevert() bool {
	if err, args := launcherCommon.RevertConfigStore.Load(); err == nil && len(args) > 0 {
		return true
	}
	return false
}

func (c *Config) revertCommand() []string {
	if err, args := launcherCommon.RevertCommandStore.Load(); err == nil {
		return args
	}
	return []string{}
}

func (c *Config) RequiresRunningRevertCommand() bool {
	return c.setupCommandRan && len(c.revertCommand()) > 0
}

func (c *Config) ServerExe() string {
	return c.serverExe
}

func (c *Config) RevertCommand() []string {
	if c.setupCommandRan {
		return c.revertCommand()
	}
	return []string{}
}

func (c *Config) HostFilePath() string {
	return c.hostFilePath
}

func (c *Config) CertFilePath() string {
	return c.certFilePath
}

func (c *Config) Revert() {
	c.KillAgent()
	if serverExe := c.ServerExe(); len(serverExe) > 0 {
		fmt.Print(printer.Gen(
			printer.Stop,
			"",
			printer.T("Stopping "),
			printer.TS("server", printer.ComponentStyle),
			printer.T("... "),
		))
		if _, err := commonProcess.Kill(serverExe); err == nil {
			printer.PrintSucceeded()
		} else {
			printer.PrintFailedError(err)
		}
	}
	if c.RequiresConfigRevert() {
		fmt.Print(
			printer.Gen(
				printer.Clean,
				"",
				printer.T("Cleaning up... "),
			),
		)
		if ok := launcherCommon.ConfigRevert(c.gameId, false, executor.RunRevert, printer.ConfigRevertPrinter()); !ok {
			printer.PrintFailed()
		} else {
			printer.PrintSimpln(printer.Success, "Cleaned.")
		}
	}
	if c.RequiresRunningRevertCommand() {
		printer.Println(
			printer.Execute,
			printer.T("Running "),
			printer.TS("Config.RevertCommand", printer.OptionStyle),
			printer.T("... "),
		)
		err := executor.RunRevertCommand()
		if err != nil {
			printer.PrintFailedError(err)
		} else {
			printer.PrintSucceeded()
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
	certStore.ReloadSystemCertificates()
	launcherCommon.ClearCache()
	return
}
