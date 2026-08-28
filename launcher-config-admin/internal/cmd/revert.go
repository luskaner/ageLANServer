package cmd

import (
	"crypto/x509"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cmd/config/admin"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
)

func trustCertificates(certificates []*x509.Certificate) bool {
	commonLogger.Println("Adding previously removed local certificate")
	if err := trustCertsFn(false, certificates); err == nil {
		commonLogger.Println("Successfully added local certificate")
		return true
	}
	commonLogger.Println("Failed to add local certificate")
	return false
}

func runRevert(args []string) (err error, exitCode int) {
	values, fs := admin.RevertFlagSet()
	if err = fs.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}
	internal.SetUp = new(false)
	if values.LogRoot != "" {
		if initErr := initializeFn(values.LogRoot); initErr != nil {
			commonLogger.Println("Failed to initialize file logging:", initErr)
		}
	}
	if values.RemoveAll {
		values.IPs = true
		values.Certs = true
	}
	var removedCertificates []*x509.Certificate
	if values.Certs {
		commonLogger.Println("Removing local certificate")
		removedCertificates, err = untrustCertsFn(false)
		if err == nil {
			commonLogger.Println("Successfully removed local certificate")
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					trustCertificates(removedCertificates)
					exitCode = common.ErrSignal
				}
			}()
		} else {
			commonLogger.Println("Failed to remove local certificate")
			commonLogger.Println("Error:", err)
			if !values.RemoveAll {
				exitCode = internal.ErrLocalCertRemove
				return
			}
		}
	}
	if values.IPs {
		commonLogger.Println("Removing IP mappings")
		if err = removeHostsFn(); err == nil {
			commonLogger.Println("Successfully removed IP mappings")
		} else {
			exitCode = internal.ErrIpMapRemove
			if !values.RemoveAll {
				if removedCertificates != nil {
					if !trustCertificates(removedCertificates) {
						exitCode = internal.ErrIpMapRemoveRevert
					}
				}
			}
			commonLogger.Println("Failed to remove IP mappings")
		}
	}
	return
}
