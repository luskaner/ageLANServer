package cmdUtils

import (
	"io"
	"path/filepath"
	"runtime"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/cmd/bsManager"
	"github.com/luskaner/ageLANServer/common/executables"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/spf13/pflag"
)

func (c *Config) RunBattleServerManager(executable string, flags *pflag.FlagSet, values *bsManager.StartValues, stop bool) (exitCode int) {
	if executable == "auto" {
		executable = executables.FindPath(executables.NativeFileName(true, "battle-server-manager"))
		if executable == "" {
			logger.Fail("Could not find battle-server-manager executable")
			return launcher.ErrBattleServerManagerRun
		}
	}
	var beforeConfigs []battleServer.Config
	if stop {
		var err error
		beforeConfigs, err = battleServer.Configs(c.gameId, true, true)
		if err != nil {
			logger.Fail("Could not get existing configurations: %s", err)
			return launcher.ErrBattleServerManagerRun
		}
	}
	if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
		values.LogRoot = logRoot
	}
	if stop {
		values.HideWindow = true
	}
	values.GameId = c.gameId
	startArgs := cmd.FlagSetToArgs(flags, true)
	str := "Running Battle-Server, "
	if runtime.GOOS != "windows" {
		str += "it can take a while and "
	}
	str += "you might need to allow it in the firewall..."
	logger.Step("%s", str)
	options := commonExecutor.Options{File: executable, Args: startArgs, Wait: true, ExitCode: true}
	var result *commonExecutor.Result
	if err := commonLogger.FileLogger.Buffer("battle-server-manager_start", func(writer io.Writer) {
		commonLogger.Println("run battle-server-manager", options.String())
		if writer != nil {
			options.Stderr = writer
			options.Stdout = writer
		}
		result = options.Exec()
	}); err != nil {
		return common.ErrFileLog
	}
	if result.Success() {
		if stop {
			afterConfigs, err := battleServer.Configs(c.gameId, true, true)
			if err == nil && len(afterConfigs) > 0 {
				if absPath, err := filepath.Abs(executable); err == nil {
					beforeConfigsSet := mapset.NewThreadUnsafeSet[battleServer.Config](beforeConfigs...)
					afterConfigsSet := mapset.NewThreadUnsafeSet[battleServer.Config](afterConfigs...)
					added := afterConfigsSet.Difference(beforeConfigsSet)
					removed := beforeConfigsSet.Difference(afterConfigsSet)
					if added.Cardinality() == 1 && removed.IsEmpty() {
						config, _ := added.Pop()
						c.battleServerRegion = config.Region
						c.battleServerExe = absPath
						return common.ErrSuccess
					}
				}
			}
			logger.Warn("A Battle Server already existed or could not determine which one was started, kill %s in task manager as needed.", battleServer.Executable)
		}
		return common.ErrSuccess
	}
	logger.Fail("Could not run battle-server-manager.")
	if result.Err != nil {
		logger.Fault("Error: %s", result.Err)
	}
	if result.ExitCode != common.ErrSuccess {
		logger.Fault("Exit code: %d", result.ExitCode)
	}
	return launcher.ErrBattleServerManagerRun
}

func (c *Config) gameRequiresBattleServer() bool {
	return c.gameId == game.AoM || c.gameId == game.AoE4
}
