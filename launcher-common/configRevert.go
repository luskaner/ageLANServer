package launcher_common

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
)

const file = common.Name + "_config_revert.txt"
const sep = "|"

var mutex sync.RWMutex

func byteToStringSlice(s []byte) []string {
	return strings.Split(string(s), sep)
}

func LoadRevertArgs() (err error, flags []string) {
	var content []byte
	mutex.RLock()
	func() {
		defer mutex.RUnlock()
		content, err = os.ReadFile(path())
	}()
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	flags = byteToStringSlice(content)
	return
}

func StoreRevertArgs(flags []string) error {
	mutex.Lock()
	defer mutex.Unlock()
	f, err := os.OpenFile(path(), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	var content []byte
	content, err = io.ReadAll(f)
	if err != nil {
		return err
	}
	flagsToSave := byteToStringSlice(content)
	existingFlags := mapset.NewSet[string](flagsToSave...)
	for _, flag := range flags {
		if !existingFlags.ContainsOne(flag) {
			flagsToSave = append(flagsToSave, flag)
		}
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = f.WriteString(strings.Join(flagsToSave, sep))
	return err
}

func RemoveRevertArgs() error {
	mutex.Lock()
	var err error
	func() {
		defer mutex.Unlock()
		err = os.Remove(path())
	}()
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func RevertFlags(game string, unmapIPs bool, removeUserCert bool, removeLocalCert bool, restoreMetadata bool, restoreProfiles bool, unmapCDN bool, hostFilePath string, certFilePath string, stopAgent bool, failfast bool) []string {
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
		if unmapCDN {
			args = append(args, "-c")
		}
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

func ConfigRevert(gameId string, headless bool, runRevertFn func(flags []string, bin bool) (result *exec.Result)) bool {
	if runRevertFn == nil {
		runRevertFn = RunRevert
	}
	err, revertFlags := LoadRevertArgs()
	if err != nil || len(revertFlags) > 0 {
		var stopAgent bool
		var revertLine string
		if !headless {
			revertLine = "Reverting "
		}
		if err != nil {
			if !headless {
				fmt.Println("Failed to get revert flags: ", err)
				revertLine += "all possible "
			}
			stopAgent = ConfigAdminAgentRunning(headless)
			revertFlags = RevertFlags(gameId, true, runtime.GOOS == "windows", true, true, true, true, "", "", stopAgent, false)
		} else if !headless && slices.Contains(revertFlags, "-g") {
			stopAgent = true
		}

		if err = RemoveRevertArgs(); err != nil {
			if !headless {
				fmt.Println("Failed to clear revert flags: ", err)
			}
		}
		requiresRevertAdminElevation := RequiresRevertAdminElevation(revertFlags, headless)
		if !headless {
			revertLine += "configuration"
			fmt.Print(revertLine)
			if requiresRevertAdminElevation {
				fmt.Print(`, authorize 'config-admin' if needed`)
			} else if stopAgent {
				fmt.Print(` and stopping its agent`)
			}
			fmt.Println(`...`)
		} else if requiresRevertAdminElevation {
			return false
		}

		if revertResult := runRevertFn(revertFlags, headless); !revertResult.Success() {
			if !headless {
				if ConfigAdminAgentRunning(false) {
					fmt.Println("'config-admin-agent' process is still executing. Kill it using the task manager with admin rights.")
				} else {
					fmt.Println("Failed to cleanup configuration, try to do it manually.")
				}
			}
			return false
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

func RequiresStopConfigAgent(args []string) bool {
	return !executor.IsAdmin() && (slices.Contains(args, "-g") || (slices.Contains(args, "-l") &&
		!slices.Contains(args, "-t")) ||
		(((slices.Contains(args, "-c")) || slices.Contains(args, "-i")) &&
			!slices.Contains(args, "-o")))
}

func RunRevert(flags []string, bin bool) (result *exec.Result) {
	args := []string{ConfigRevertCmd}
	args = append(args, flags...)
	result = exec.Options{File: common.GetExeFileName(bin, common.LauncherConfig), Wait: true, Args: args, ExitCode: true}.Exec()
	return
}

func path() string {
	return filepath.Join(os.TempDir(), file)
}
