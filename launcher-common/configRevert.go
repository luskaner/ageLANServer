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

func RevertFlags(game string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreGameCert bool, restoreMetadata bool, restoreProfiles bool, hostFilePath string, certFilePath string, gamePath string, logRoot string, stopAgent bool, failfast bool) []string {
	args := make([]string, 0)
	if game != "" {
		args = append(args, "-e")
		args = append(args, game)
	}
	if stopAgent {
		args = append(args, "-g")
	}
	if !failfast {
		args = append(args, "-a")
	} else {
		if unmapIPs {
			args = append(args, "-i")
		}
		if removeUserCert {
			args = append(args, "-u")
		}
		if removeLocalCert {
			args = append(args, "-l")
		}
		if restoreMetadata {
			args = append(args, "-m")
		}
		if restoreProfiles {
			args = append(args, "-p")
		}
		if restoreGameCert {
			args = append(args, "-s")
		}
	}
	if gamePath != "" {
		args = append(args, "--gamePath")
		args = append(args, gamePath)
	}
	if logRoot != "" {
		args = append(args, "--logRoot")
		args = append(args, logRoot)
	}
	if hostFilePath != "" {
		args = append(args, "-o")
		args = append(args, hostFilePath)
	}
	if certFilePath != "" {
		args = append(args, "-t")
		args = append(args, certFilePath)
	}
	return args
}

func allRevertFlags(gameId string, logRoot string, stopAgent bool) []string {
	return RevertFlags(gameId, true, runtime.GOOS == "windows", true, false, true, true, "", "", "", logRoot, stopAgent, false)
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
			commonLogger.Printf("Failed to get revert flags: %v, will revert for all games\n", err)
			stopAgent = ConfigAdminAgentRunning(headless)
			for i, game := range common.SupportedGames.ToSlice() {
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
		for i, currentRevertFlags := range multipleRevertFlags {
			commonLogger.Println(games[i] + ":")
			commonLogger.Println("\tReverting configuration" + revertEnd + `...`)
			if revertResult := runRevertFn(currentRevertFlags, headless, out, optionsFn); revertResult.Success() {
				success = true
			} else {
				if ConfigAdminAgentRunning(false) {
					commonLogger.Println("\t\t'config-admin-agent' process is still executing. Kill it using the task manager with admin rights.")
				} else {
					commonLogger.Println("\t\tFailed to cleanup configuration, try to do it manually.")
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
	if _, proc, err := commonProcess.Process(executables.Filename(bin, executables.LauncherConfigAdminAgent)); err == nil && proc != nil {
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
	options := exec.Options{File: executables.Filename(bin, executables.LauncherConfig), Wait: true, Args: args, ExitCode: true}
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
