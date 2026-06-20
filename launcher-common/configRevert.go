package launcher_common

import (
	"io"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/spf13/pflag"
)

var RevertConfigStore = NewArgsStore(filepath.Join(os.TempDir(), common.Name+"_config_revert.txt"))

type ConfigRevertFlagOptions struct {
	*config.RevertValues
	flags *pflag.FlagSet
}

func NewConfigRevertFlagOptions() *ConfigRevertFlagOptions {
	values, flags := config.RevertFlagSet()
	return &ConfigRevertFlagOptions{
		RevertValues: values,
		flags:        flags,
	}
}

func (c *ConfigRevertFlagOptions) Flags() []string {
	if c.RemoveAll {
		c.IPs = false
		c.RemoveUserCert = false
		c.Certs = false
		c.Metadata = false
		c.Profiles = false
		c.RestoreCAStoreCert = false
	}
	return commonCmd.FlagSetToArgs(c.flags, false)
}

func allRevertFlags(gameId string, logRoot string) []string {
	options := NewConfigRevertFlagOptions()
	options.GameId = gameId
	options.LogRoot = logRoot
	options.RemoveAll = true
	return options.Flags()
}

func ConfigRevert(
	gameId string,
	logRoot string,
	headless bool,
	out io.Writer,
	optionsFn func(options exec.Options),
	runRevertFn func(flags []string, bin bool, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result),
) (success bool) {
	if runRevertFn == nil {
		runRevertFn = RunRevert
	}
	err, revertFlags := RevertConfigStore.Load()
	var games []string
	if gameId == "" {
		games = game.SupportedGames.ToSlice()
	} else {
		games = []string{gameId}
	}
	multipleRevertFlags := make([][]string, len(games))
	if err != nil || len(revertFlags) > 0 {
		if err == nil {
			_, flags := config.RevertFlagSet()
			if err = flags.Parse(revertFlags); err != nil {
				commonLogger.Printf("Failed to parse revert flags: %v\n", err)
			} else {
				multipleRevertFlags = [][]string{commonCmd.FlagSetToArgs(flags, false)}
			}
		}
		if err != nil {
			if len(games) == 1 {
				commonLogger.Printf("Failed to get revert flags: %v, will revert for game %s\n", err, games[0])
			} else {
				commonLogger.Printf("Failed to get revert flags: %v, will revert for all games\n", err)
			}
			for i, g := range games {
				multipleRevertFlags[i] = allRevertFlags(g, logRoot)
			}
		}
		// This does not depend on the game type so compute it once
		requiresRevertAdminElevation := RevertRequiresAdminElevation(multipleRevertFlags[0], headless)
		if headless && requiresRevertAdminElevation {
			commonLogger.Println("Revert requires admin elevation while headless, this should not happen, skipping...")
			return
		}
		var revertEnd string
		if requiresRevertAdminElevation {
			revertEnd += `, authorize 'config-admin' if needed`
		}
		for _, currentRevertFlags := range multipleRevertFlags {
			commonLogger.Println("Reverting configuration" + revertEnd + `...`)
			if revertResult := runRevertFn(currentRevertFlags, headless, out, optionsFn); revertResult.Success() {
				success = true
			} else {
				if ConfigAdminAgentRunning(headless) {
					commonLogger.Println("\t'config-admin-agent' process is still executing. Kill it using the task manager with admin rights.")
				} else {
					commonLogger.Println("\tFailed to cleanup configuration, try to do it manually.")
				}
			}
		}
		if success {
			if err = RevertConfigStore.Delete(); err != nil {
				commonLogger.Println("Failed to clear revert flags: ", err)
			}
		}
	} else {
		success = true
	}
	return success
}

func ConfigAdminAgentRunning(bin bool) bool {
	if _, proc, err := commonProcess.Process(executables.NativeFileName(bin, executables.LauncherConfigAdminAgent)); err == nil && proc != nil {
		return true
	}
	return false
}

func RequiresAdminElevation(bin bool) bool {
	return !executor.IsAdmin() && !ConfigAdminAgentRunning(bin)
}

func RevertRequiresAdminElevation(args []string, bin bool) bool {
	if !RequiresAdminElevation(bin) {
		return false
	}
	values, flags := config.RevertFlagSet()
	// If there is an error parsing the args assume worst-case scenario, admin is needed.
	if err := flags.Parse(args); err != nil {
		commonLogger.Println("Failed to parse revert flags: ", err, ", assuming admin elevation is needed")
		return true
	}
	return RevertRequiresAdminElevationValues(values)
}

func RevertRequiresAdminElevationValues(values *config.RevertValues) bool {
	return (values.Certs && values.CertFilePath == "") ||
		(values.IPs && values.HostFilePath == "")
}

func RunRevert(flags []string, bin bool, out io.Writer, optionsFn func(options exec.Options)) (result *exec.Result) {
	args := []string{ConfigRevertCmd}
	args = append(args, flags...)
	options := exec.Options{File: executables.NativeFileName(bin, executables.LauncherConfig), Wait: true, Args: args, ExitCode: true}
	if optionsFn != nil {
		optionsFn(options)
	}
	if out != nil {
		options.Stdout = out
		options.Stderr = out
	}
	result = options.Exec()
	return
}
