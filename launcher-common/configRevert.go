package launcher_common

import (
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

type ConfigRevertPrinter struct {
	Revert              func(all bool, requiresRevertAdminElevation bool, stopAgent bool)
	RevertFlagsErr      func(err error)
	ClearRevertFlagsErr func(err error)
	RevertResult        func(result *exec.Result)
}

func stubConfigRevertPrinter() *ConfigRevertPrinter {
	return &ConfigRevertPrinter{
		Revert:              func(_, _, _ bool) {},
		RevertFlagsErr:      func(_ error) {},
		ClearRevertFlagsErr: func(_ error) {},
		RevertResult:        func(_ *exec.Result) {},
	}
}

func ConfigRevert(
	gameId string,
	binCannotElevate bool,
	runRevertFn func(flags []string, bin bool) (result *exec.Result),
	printer *ConfigRevertPrinter,
) bool {
	if runRevertFn == nil {
		runRevertFn = RunRevert
	}
	if printer == nil {
		printer = stubConfigRevertPrinter()
	}
	err, revertFlags := RevertConfigStore.Load()
	var stopAgent bool
	if _, _, errAgent := ConfigAdminAgent(binCannotElevate); errAgent == nil {
		stopAgent = true
	}
	allRevertFlags := func(stopAgent bool) []string {
		return RevertFlags(&RevertFlagsOptions{
			Game:            gameId,
			UnmapIP:         true,
			UnmapCDN:        true,
			RemoveUserCert:  runtime.GOOS == "windows",
			RemoveLocalCert: true,
			RestoreMetadata: true,
			StopAgent:       stopAgent,
		})
	}
	if err != nil || (len(revertFlags) == 0 && stopAgent) {
		revertFlags = allRevertFlags(stopAgent)
	}
	if len(revertFlags) > 0 {
		requiresRevertAdminElevation := RequiresRevertAdminElevation(revertFlags, binCannotElevate)
		if binCannotElevate && requiresRevertAdminElevation {
			return false
		}
		var all bool
		if err != nil {
			all = true
			printer.RevertFlagsErr(err)
		}
		if err = RevertConfigStore.Delete(); err != nil {
			printer.ClearRevertFlagsErr(err)
		}
		printer.Revert(all, requiresRevertAdminElevation, stopAgent)
		if stopAgent && !slices.Contains(revertFlags, "-g") {
			revertFlags = append(revertFlags, "-g")
		}
		result := runRevertFn(revertFlags, binCannotElevate)
		printer.RevertResult(result)
		return result.Success()
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
