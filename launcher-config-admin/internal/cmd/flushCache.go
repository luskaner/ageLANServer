package cmd

import (
	"github.com/luskaner/ageLANServer/common"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
)

func runFlushCache(args []string) (err error, exitCode int) {
	values, fs := config.FlushCacheFlagSet()
	if err = fs.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	if values.LogRoot != "" {
		if initErr := initializeFn(values.LogRoot); initErr != nil {
			commonLogger.Println("Failed to initialize file logging:", initErr)
		}
	}
	if values.Certs {
		if runtimeGOOS != "windows" {
			commonLogger.Println("Flushing Certs cache...")
			if result := flushCertsFn(); !result.Success() {
				commonLogger.Println("Failed to flush Certs cache")
				if result.ExitCode != common.ErrSuccess {
					commonLogger.Printf("Exit code: %v\n", result.ExitCode)
				}
				if result.Err != nil {
					commonLogger.Printf("Error: %v\n", result.Err)
				}
				exitCode = internal.ErrFlushCacheCerts
			}
		}
	}
	if values.IPs {
		if result := flushDnsFn(); !result.Success() {
			commonLogger.Println("Failed to flush DNS cache")
			if result.ExitCode != common.ErrSuccess {
				commonLogger.Printf("Exit code: %v\n", result.ExitCode)
			}
			if result.Err != nil {
				commonLogger.Printf("Error: %v\n", result.Err)
			}
			if exitCode == internal.ErrFlushCacheCerts {
				exitCode = internal.ErrFlushCache
			} else {
				exitCode = internal.ErrFlushCacheDNS
			}
		}
	}
	return
}
