package hosts

import (
	"io"

	"github.com/hairyhenderson/go-which"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor/exec"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
)

func FlushDns() (result *exec.Result) {
	path := which.Which("resolvectl")
	if path != "" {
		commonLogger.Println("Flushing DNS cache...")
		options := exec.Options{File: path, ExitCode: true, Wait: true, Args: []string{"flush-caches"}}
		var suffix string
		if internal.SetUp == nil {
			suffix = "_flushCache"
		} else if *internal.SetUp {
			suffix = "_setup"
		} else {
			suffix = "_revert"
		}
		if err := internal.Logger.Buffer("resolvectl_flushcache"+suffix, func(writer io.Writer) {
			if writer != nil {
				options.Stdout = writer
				options.Stderr = writer
				commonLogger.Printf("run resolvectl: %s\n", options.String())
			}
			result = options.Exec()
		}); err != nil {
			result = &exec.Result{
				ExitCode: common.ErrFileLog,
				Err:      err,
			}
		}
		return
	}
	// Some systems do not have, assume it is not needed
	return exec.ResultSuccess
}
