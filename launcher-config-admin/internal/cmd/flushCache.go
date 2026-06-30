package cmd

import (
	"runtime"

	"github.com/luskaner/ageLANServer/common"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
)

func runFlushCache(args []string) (err error, exitCode int) {
	values, fs := config.FlushCacheFlagSet()
	if err = fs.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	if values.LogRoot != "" {
		internal.Initialize(values.LogRoot)
	}
	if values.Certs {
		if runtime.GOOS != "windows" {
			commonLogger.Println("Flushing Certs cache...")
			if result := cert.FlushCerts(); !result.Success() {
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
		if result := hosts.FlushDns(); !result.Success() {
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
