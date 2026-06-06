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
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
	"github.com/spf13/pflag"
)

func (c *Config) RunBattleServerManager(executable string, flags *pflag.FlagSet, values *bsManager.StartValues, stop bool) (errorCode int) {
	if executable == "auto" {
		executable = executables.FindPath(executables.NativeFileName(true, "battle-server-manager"))
		if executable == "" {
			logger.Println("Could not find 'battle-server-manager' executable")
			return internal.ErrBattleServerManagerRun
		}
	}
	var beforeConfigs []battleServer.Config
	if stop {
		var err error
		beforeConfigs, err = battleServer.Configs(c.gameId, true)
		if err != nil {
			logger.Println("Could not get existing configurations:", err)
			return internal.ErrBattleServerManagerRun
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
	str := "Running 'battle-server-manager, "
	if runtime.GOOS != "windows" {
		str += "it can take a while and "
	}
	str += "you might need to allow it in the firewall..."
	logger.Println(str)
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
			afterConfigs, err := battleServer.Configs(c.gameId, true)
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
			logger.Printf("A Battle Server already existed or could not determine which one was started, kill '%s' in task manager as needed.\n", battleServer.Executable)
		}
		return common.ErrSuccess
	}
	logger.Println("Could not run 'battle-server-manager'.")
	if result.Err != nil {
		logger.Println("Error:", result.Err)
	}
	if result.ExitCode != common.ErrSuccess {
		logger.Println("Exit code:", result.ExitCode)
	}
	return internal.ErrBattleServerManagerRun
}
