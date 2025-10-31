package cmdUtils

import (
	"io"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/cmdUtils/logger"
)

func (c *Config) RunBattleServerManager(executable string, args []string, stop bool) (errorCode int) {
	if executable == "auto" {
		executable = common.FindExecutablePath(common.GetExeFileName(true, "battle-server-manager"))
		if executable == "" {
			logger.Println("Could not find 'battle-server-manager' executable")
			return internal.ErrBattleServerManagerRun
		}
	}
	var beforeConfigs []battleServerConfig.Config
	if stop {
		var err error
		beforeConfigs, err = battleServerConfig.Configs(c.gameId, true)
		if err != nil {
			logger.Println("Could not get existing configurations:", err)
			return internal.ErrBattleServerManagerRun
		}
	}
	finalArgs := []string{"start"}
	if stop {
		finalArgs = append(finalArgs, "-w")
	}
	if logRoot := commonLogger.FileLogger.Folder(); logRoot != "" {
		args = append(args, "--logRoot", logRoot)
	}
	finalArgs = append(finalArgs, args...)
	logger.Println("Running 'battle-server-manager', you might to allow it in the firewall...")
	options := commonExecutor.Options{File: executable, Args: finalArgs, Wait: true, ExitCode: true}
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
			afterConfigs, err := battleServerConfig.Configs(c.gameId, true)
			if err == nil && len(afterConfigs) > 0 {
				if absPath, err := filepath.Abs(executable); err == nil {
					beforeConfigsSet := mapset.NewThreadUnsafeSet[battleServerConfig.Config](beforeConfigs...)
					afterConfigsSet := mapset.NewThreadUnsafeSet[battleServerConfig.Config](afterConfigs...)
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
			logger.Println("A Battle Server already existed or could not determine which one was started, kill 'BattleServer.exe' in task manager as needed.")
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
