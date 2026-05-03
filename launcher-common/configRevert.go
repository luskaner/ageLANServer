package launcher_common

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
)

var RevertConfigStore = NewArgsStore(filepath.Join(os.TempDir(), common.Name+"_config_revert.txt"))

type ConfigRevertFlagOptions struct {
	GameId          string
	UnmapIPs        bool
	RemoveUserCert  bool
	RemoveLocalCert bool
	RestoreGameCert bool
	RestoreMetadata bool
	RestoreProfiles bool
	HostFilePath    string
	CertFilePath    string
	GameBinPath     string
	GameDataPath    string
	LogRoot         string
	StopAgent       bool
	FailFast        bool
}

func (c *ConfigRevertFlagOptions) Flags() []string {
	args := make([]string, 0)
	if c.GameId != "" {
		args = append(args, "-e")
		args = append(args, c.GameId)
	}
	if c.StopAgent {
		args = append(args, "-g")
	}
	if !c.FailFast {
		args = append(args, "-a")
	} else {
		if c.UnmapIPs {
			args = append(args, "-i")
		}
		if c.RemoveUserCert {
			args = append(args, "-u")
		}
		if c.RemoveLocalCert {
			args = append(args, "-l")
		}
		if c.RestoreMetadata {
			args = append(args, "-m")
		}
		if c.RestoreProfiles {
			args = append(args, "-p")
		}
		if c.RestoreGameCert {
			args = append(args, "-s")
		}
	}
	if c.GameBinPath != "" {
		args = append(args, "--gamePath")
		args = append(args, c.GameBinPath)
	}
	if c.GameDataPath != "" {
		args = append(args, "--dataPath")
		args = append(args, c.GameDataPath)
	}
	if c.LogRoot != "" {
		args = append(args, "--logRoot")
		args = append(args, c.LogRoot)
	}
	if c.HostFilePath != "" {
		args = append(args, "-o")
		args = append(args, c.HostFilePath)
	}
	if c.CertFilePath != "" {
		args = append(args, "-t")
		args = append(args, c.CertFilePath)
	}
	return args
}

func allRevertFlags(gameId string, logRoot string, stopAgent bool) []string {
	options := &ConfigRevertFlagOptions{
		GameId:          gameId,
		UnmapIPs:        true,
		RemoveUserCert:  runtime.GOOS != "linux",
		RemoveLocalCert: true,
		LogRoot:         logRoot,
		StopAgent:       stopAgent,
	}
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
		games = common.SupportedGames.ToSlice()
	} else {
		games = []string{gameId}
	}
	multipleRevertFlags := make([][]string, len(games))
	if err != nil || len(revertFlags) > 0 {
		var stopAgent bool
		if err != nil {
			if len(games) == 1 {
				commonLogger.Printf("Failed to get revert flags: %v, will revert for game %s\n", err, games[0])
			} else {
				commonLogger.Printf("Failed to get revert flags: %v, will revert for all games\n", err)
			}
			stopAgent = ConfigAdminAgentRunning(headless)
			for i, game := range games {
				multipleRevertFlags[i] = allRevertFlags(game, logRoot, stopAgent)
			}
		} else {
			multipleRevertFlags = [][]string{revertFlags}
			if !headless && slices.Contains(revertFlags, "-g") {
				stopAgent = true
			}
		}
		// This does not depend on the game type so compute it once
		requiresRevertAdminElevation := RequiresRevertAdminElevation(multipleRevertFlags[0], headless)
		if headless && requiresRevertAdminElevation {
			commonLogger.Println("Revert requires admin elevation while headless, this should not happen, skipping...")
			return
		}
		var revertEnd string
		if requiresRevertAdminElevation {
			revertEnd += `, authorize 'config-admin' if needed`
		} else if stopAgent {
			revertEnd += ` and stopping its agent`
		}
		for _, currentRevertFlags := range multipleRevertFlags {
			commonLogger.Println("Reverting configuration" + revertEnd + `...`)
			if revertResult := runRevertFn(currentRevertFlags, headless, out, optionsFn); revertResult.Success() {
				success = true
			} else {
				if ConfigAdminAgentRunning(false) {
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

func RequiresRevertAdminElevation(args []string, bin bool) bool {
	if executor.IsAdmin() {
		return false
	}
	if ConfigAdminAgentRunning(bin) {
		return false
	}
	if (slices.Contains(args, "-l") &&
		!slices.Contains(args, "-t")) ||
		(((slices.Contains(args, "-c")) || slices.Contains(args, "-i")) &&
			!slices.Contains(args, "-o")) {
		return true
	}
	return false
}

func RequiresStopConfigAgent(args []string) bool {
	return !executor.IsAdmin() && (slices.Contains(args, "-g") || (slices.Contains(args, "-l") &&
		!slices.Contains(args, "-t")) ||
		(((slices.Contains(args, "-c")) || slices.Contains(args, "-i")) &&
			!slices.Contains(args, "-o")))
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
