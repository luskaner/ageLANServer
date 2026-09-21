package cmd

import (
	"crypto/x509"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	launcherCommonHosts "github.com/luskaner/ageLANServer/common/hosts"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config/admin"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
)

func untrustCertificate() bool {
	commonLogger.Println("Removing previously added local certificate")
	if _, err := untrustCertsFn(false); err == nil {
		commonLogger.Println("Successfully removed local certificate")
		return true
	}
	commonLogger.Println("Failed to remove local certificate")
	return false
}

func logHostsDiagnostics() {
	hostsPath := launcherCommonHosts.Path()
	commonLogger.Printf("Hosts file path: %q\n", hostsPath)
	if info, statErr := os.Stat(hostsPath); statErr != nil {
		commonLogger.Printf("Hosts file stat failed: %v\n", statErr)
	} else {
		commonLogger.Printf("Hosts file info: size=%d mode=%v isDir=%v\n", info.Size(), info.Mode(), info.IsDir())
	}
	backupPath := filepath.Join(filepath.Dir(hostsPath), "hosts.bak")
	commonLogger.Printf("Hosts backup path: %q\n", backupPath)
	if info, statErr := os.Stat(backupPath); statErr != nil {
		commonLogger.Printf("Hosts backup stat: %v\n", statErr)
	} else {
		commonLogger.Printf("Hosts backup info: size=%d mode=%v isDir=%v\n", info.Size(), info.Mode(), info.IsDir())
		commonLogger.Println("If the backup already exists from a previous failed run, delete it after verifying its contents and retry.")
	}
}

func runSetUp(args []string) (err error, exitCode int) {
	values, fs := admin.SetupFlagSet()
	if err = fs.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}

	// validate required flags
	if values.GameId == "" {
		return errors.New("required flag 'game' not set"), common.ErrSyntax
	}

	internal.SetUp = new(true)
	if values.LogRoot != "" {
		if initErr := initializeFn(values.LogRoot); initErr != nil {
			commonLogger.Println("Failed to initialize file logging:", initErr)
		}
	}
	trustedCertificate := false
	if len(values.AddLocalCertData) > 0 {
		commonLogger.Println("Adding local certificate")
		crt := bytesToCertFn(values.AddLocalCertData)
		if crt == nil {
			commonLogger.Println("Failed to parse certificate")
			exitCode = internal.ErrLocalCertAddParse
			return
		}
		if err = trustCertsFn(false, []*x509.Certificate{crt}); err == nil {
			commonLogger.Println("Successfully added local certificate")
			trustedCertificate = true
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					untrustCertificate()
					exitCode = common.ErrSignal
				}
			}()
		} else {
			commonLogger.Println("Failed to add local certificate")
			commonLogger.Println("Error:", err)
			exitCode = internal.ErrLocalCertAdd
			return
		}
	}
	if len(values.MapIp) > 0 {
		commonLogger.Printf("Adding IP mappings for game %q with IP %q...\n", values.GameId, values.MapIp.String())
		if ok, addHostsErr := addHostsFn(values.MapIp, values.GameId, "", "", values.MacOsExclusiveMappings, hosts.FlushDns); ok {
			commonLogger.Println("Successfully added IP mappings")
		} else {
			exitCode = internal.ErrIpMapAdd
			commonLogger.Println("Failed to add IP mappings")
			if addHostsErr != nil {
				commonLogger.Println("Error message:", addHostsErr)
			} else {
				commonLogger.Println("Error message: unknown (AddHosts returned ok=false with nil error)")
			}
			logHostsDiagnostics()
			if trustedCertificate {
				if !untrustCertificate() {
					exitCode = internal.ErrIpMapAddRevert
				}
			}
		}
	}
	return
}
