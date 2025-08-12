package cmdUtils

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
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
	gameTitle       common.GameTitle
	serverPid       int
	setupCommandRan bool
	hostFilePath    string
	certFilePath    string
	ipProtocol      common.IPProtocol
}

func (c *Config) SetIPProtocol(ipProtocol common.IPProtocol) {
	c.ipProtocol = ipProtocol
}

func (c *Config) SetServerPid(pid int) {
	c.serverPid = pid
}

func (c *Config) SetGameTitle(gameTitle common.GameTitle) {
	c.gameTitle = gameTitle
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

func (c *Config) IPProtocol() *common.IPProtocol {
	return &c.ipProtocol
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

func (c *Config) ServerPid() int {
	return c.serverPid
}

func (c *Config) GameTitle() common.GameTitle {
	return c.gameTitle
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
	if serverPid := c.ServerPid(); serverPid != 0 {
		fmt.Print(printer.Gen(
			printer.Stop,
			"",
			printer.T("Stopping "),
			printer.TS("server", printer.ComponentStyle),
			printer.T("... "),
		))
		if err := commonProcess.KillPid(serverPid); err == nil {
			printer.PrintSucceeded()
		} else {
			printer.PrintFailedError(err)
		}
	}
	if c.RequiresConfigRevert() {
		printer.PrintSimpln(
			printer.Clean,
			"Cleaning up... ",
		)
		if ok := launcherCommon.ConfigRevert(c.GameTitle(), false, executor.RunRevert, printer.ConfigRevertPrinter()); !ok {
			printer.PrintFailed()
		} else {
			printer.PrintSimpln(printer.Success, "Cleaned.")
		}
	}
	if c.RequiresRunningRevertCommand() {
		printer.Println(
			printer.Execute,
			printer.T("Running "),
			printer.TS("RevertCommand", printer.OptionStyle),
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

func GamesRunning() mapset.Set[common.GameTitle] {
	games := mapset.NewThreadUnsafeSet[common.GameTitle]()
	xbox := runtime.GOOS == "windows"
	for gameTitle := range common.AllGameTitles.Iter() {
		if anyProcessExists(commonProcess.GameProcesses(gameTitle, true, xbox)) {
			games.Add(gameTitle)
		}
	}
	return games
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
