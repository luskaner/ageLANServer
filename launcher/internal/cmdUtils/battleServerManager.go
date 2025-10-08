package cmdUtils

import (
	"fmt"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/launcher/internal"
)

func (c *Config) RunBattleServerManager(gameId string, executable string, args []string, stop bool) (errorCode int) {
	if executable == "auto" {
		executable = common.FindExecutablePath(common.GetExeFileName(true, "battle-server-manager"))
		if executable == "" {
			fmt.Println("Could not find 'battle-server-manager' executable")
			return internal.ErrBattleServerManagerRun
		}
	}
	var beforeConfigs []battleServerConfig.Config
	if stop {
		var err error
		beforeConfigs, err = battleServerConfig.Configs(gameId, true)
		if err != nil {
			fmt.Println("Could not get existing configurations:", err)
			return internal.ErrBattleServerManagerRun
		}
	}
	finalArgs := []string{"start"}
	if stop {
		finalArgs = append(finalArgs, "-w")
	}
	finalArgs = append(finalArgs, args...)
	fmt.Println("Running 'battle-server-manager', you might to allow it in the firewall...")
	result := commonExecutor.Options{File: executable, Args: finalArgs, Wait: true, ExitCode: true}.Exec()
	if result.Success() {
		if stop {
			afterConfigs, err := battleServerConfig.Configs(gameId, true)
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
			fmt.Println("A Battle Server already existed or could not determine which one was started, kill 'BattleServer.exe' in task manager as needed.")
		}
		return common.ErrSuccess
	}
	fmt.Println("Could not run 'battle-server-manager'.")
	if result.Err != nil {
		fmt.Println("Error:", result.Err)
	}
	if result.ExitCode != common.ErrSuccess {
		fmt.Println("Exit code:", result.ExitCode)
	}
	return internal.ErrBattleServerManagerRun
}
