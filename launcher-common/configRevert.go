package launcher_common

import (
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"os"
	"path/filepath"
	"runtime"
	"slices"
)

var RevertConfigStore = NewArgsStore(filepath.Join(os.TempDir(), common.Name+"_config_revert.txt"))

type RevertFlagsOptions struct {
	Game                   string
	HostFilePath           string
	UnmapIPs               bool
	UnmapCDN               bool
	CertFilePath           string
	RemoveUserCert         bool
	RemoveLocalCert        bool
	WindowsUserProfilePath string
	RestoreMetadata        bool
	RestoreProfiles        bool
	StopAgent              bool
	Failfast               bool
}

func RevertFlags(options *RevertFlagsOptions) []string {
	args := make([]string, 0)
	if options.Game != "" {
		args = append(args, "-e")
		args = append(args, options.Game)
	}
	if options.StopAgent {
		args = append(args, "-g")
	}
	if !options.Failfast {
		args = append(args, "-a")
	} else {
		if options.UnmapIPs {
			args = append(args, "-i")
		}
		if options.RemoveUserCert {
			args = append(args, "-u")
		}
		if options.RemoveLocalCert {
			args = append(args, "-l")
		}
		if options.RestoreMetadata {
			args = append(args, "-m")
		}
		if options.RestoreProfiles {
			args = append(args, "-p")
		}
		if options.UnmapCDN {
			args = append(args, "-c")
		}
	}
	if options.HostFilePath != "" {
		args = append(args, "-o")
		args = append(args, options.HostFilePath)
	}
	if options.CertFilePath != "" {
		args = append(args, "-t")
		args = append(args, options.CertFilePath)
	}
	if options.WindowsUserProfilePath != "" {
		args = append(args, "-s")
		args = append(args, options.WindowsUserProfilePath)
	}
	return args
}

func ConfigRevert(gameId string, headless bool, runRevertFn func(flags []string, bin bool) (result *exec.Result)) bool {
	if runRevertFn == nil {
		runRevertFn = RunRevert
	}
	err, revertFlags := RevertConfigStore.Load()
	stopAgent := ConfigAdminAgentRunning(headless)
	var revertLine string

	if err != nil || (len(revertFlags) == 0 && stopAgent) {
		revertFlags = RevertFlags(&RevertFlagsOptions{
			Game:            gameId,
			UnmapIPs:        true,
			UnmapCDN:        true,
			RemoveUserCert:  runtime.GOOS == "windows",
			RemoveLocalCert: true,
			RestoreMetadata: true,
			RestoreProfiles: true,
			StopAgent:       stopAgent,
		})
	}
	if len(revertFlags) > 0 {
		requiresRevertAdminElevation := RequiresRevertAdminElevation(revertFlags, headless)
		if headless && requiresRevertAdminElevation {
			return false
		}
		if !headless {
			revertLine = "Reverting "
		}
		if err != nil && !headless {
			fmt.Println("Failed to get revert flags: ", err)
			revertLine += "all possible "
		}
		if err = RevertConfigStore.Delete(); err != nil && !headless {
			fmt.Println("Failed to clear revert flags: ", err)
		}

		if !headless {
			revertLine += "configuration"
			fmt.Print(revertLine)
			if requiresRevertAdminElevation {
				fmt.Print(`, authorize 'config-admin' if needed`)
			} else if stopAgent {
				fmt.Print(` and stopping its agent`)
			}
			fmt.Println(`...`)
		}

		if stopAgent && !slices.Contains(revertFlags, "-g") {
			revertFlags = append(revertFlags, "-g")
		}
		success := runRevertFn(revertFlags, headless).Success()
		if !success && !headless {
			fmt.Println("Failed to cleanup configuration, try to do it manually.")
		}
		if ConfigAdminAgentRunning(false) {
			if !headless {
				fmt.Println("'config-admin-agent' process is still executing. Kill it using the task manager with admin rights.")
			}
			return false
		} else {
			return success
		}
	}
	return true
}

func ConfigAdminAgentRunning(bin bool) bool {
	if _, proc, err := commonProcess.Process(common.GetExeFileName(bin, common.LauncherConfigAdminAgent)); err == nil && proc != nil {
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

func RunRevert(flags []string, bin bool) (result *exec.Result) {
	args := []string{ConfigRevertCmd}
	args = append(args, flags...)
	result = exec.Options{File: common.GetExeFileName(bin, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	return
}
