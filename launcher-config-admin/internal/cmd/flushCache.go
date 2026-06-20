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

func runFlushCache(args []string) error {
	values, fs := config.FlushCacheFlagSet()
	if err := fs.Parse(args); err != nil {
		return err
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
		}
	}
	return nil
}
