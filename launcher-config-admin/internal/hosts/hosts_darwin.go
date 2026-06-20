package hosts

import (
	"io"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
)

func FlushDns() (result *exec.Result) {
	commonLogger.Println("Flushing DNS cache...")
	options := exec.Options{
		File:        "dscacheutil",
		SpecialFile: true,
		ExitCode:    true,
		Wait:        true,
		Args:        []string{"-flushcache"},
	}
	var suffix string
	if internal.SetUp == nil {
		suffix = "_flushCache"
	} else if *internal.SetUp {
		suffix = "_setup"
	} else {
		suffix = "_revert"
	}
	if err := internal.Logger.Buffer("dscacheutil_flushcache"+suffix, func(writer io.Writer) {
		if writer != nil {
			options.Stdout = writer
			options.Stderr = writer
			commonLogger.Printf("run dscacheutil: %s\n", options.String())
		}
		result = options.Exec()
	}); err != nil {
		result = &exec.Result{ExitCode: common.ErrFileLog, Err: err}
		return result
	}
	options = exec.Options{
		File:        "killall",
		SpecialFile: true,
		ExitCode:    true,
		Wait:        true,
		Args:        []string{"-HUP", "mDNSResponder"},
	}
	if err := internal.Logger.Buffer("killall_hup_mDNSResponder"+suffix, func(writer io.Writer) {
		if writer != nil {
			options.Stdout = writer
			options.Stderr = writer
			commonLogger.Printf("run killall: %s\n", options.String())
		}
		result = options.Exec()
	}); err != nil {
		result = &exec.Result{ExitCode: common.ErrFileLog, Err: err}
	}
	return
}
