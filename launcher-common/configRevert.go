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
	UnmapIP                bool
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
		if options.UnmapIP {
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
	var stopAgent bool
	if _, _, errAgent := ConfigAdminAgent(headless); errAgent == nil {
		stopAgent = true
	}
	var revertLine string
	allRevertFlags := func(stopAgent bool) []string {
		return RevertFlags(&RevertFlagsOptions{
			Game:            gameId,
			UnmapIP:         true,
			UnmapCDN:        true,
			RemoveUserCert:  runtime.GOOS == "windows",
			RemoveLocalCert: true,
			RestoreMetadata: true,
			RestoreProfiles: true,
			StopAgent:       stopAgent,
		})
	}

	if err != nil || (len(revertFlags) == 0 && stopAgent) {
		revertFlags = allRevertFlags(stopAgent)
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
			fmt.Print("Failed to cleanup configuration")
			if !executor.IsAdmin() {
				fmt.Print(", try to do it manually")
			}
			fmt.Println(".")
		}
		var pidPath string
		var proc *os.Process
		if pidPath, proc, err = ConfigAdminAgent(false); err == nil {
			if executor.IsAdmin() {
				if !headless {
					fmt.Println("Killing 'config-admin-agent' process...")
				}
				if err = commonProcess.KillProc(pidPath, proc); err == nil {
					if !headless {
						fmt.Println("Successfully killed 'config-admin-agent' process.")
						fmt.Println("Trying to revert all configuration...")
					}
					success = runRevertFn(allRevertFlags(false), headless).Success()
					if success {
						return true
					}
					if !headless {
						fmt.Println("Failed to cleanup configuration even as admin, try to do it manually.")
					}
				} else {
					if !headless {
						fmt.Println("Failed to kill 'config-admin-agent' process: ", err)
					}
				}
			} else if !headless {
				fmt.Println("'config-admin-agent' process is still executing. Kill it using the task manager with admin rights.")
			}
			return false
		} else {
			return success
		}
	}
	return true
}

func ConfigAdminAgent(bin bool) (pidPath string, proc *os.Process, err error) {
	return commonProcess.Process(common.GetExeFileName(bin, common.LauncherConfigAdminAgent))
}

func RequiresRevertAdminElevation(args []string, bin bool) bool {
	if executor.IsAdmin() {
		return false
	}
	if _, _, err := ConfigAdminAgent(bin); err == nil {
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
