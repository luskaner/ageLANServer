package cmd

import (
	"crypto/x509"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config/admin"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
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
		commonLogger.Println("Adding IP mappings")
		if ok, _ := addHostsFn(values.MapIp, values.GameId, "", "", values.MacOsExclusiveMappings, flushDnsFn); ok {
			commonLogger.Println("Successfully added IP mappings")
		} else {
			exitCode = internal.ErrIpMapAdd
			if trustedCertificate {
				if !untrustCertificate() {
					exitCode = internal.ErrIpMapAddRevert
				}
			}
			commonLogger.Println("Failed to add IP mappings")
		}
	}
	return
}
