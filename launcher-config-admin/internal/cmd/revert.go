package cmd

import (
	"crypto/x509"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
)

func trustCertificates(certificates []*x509.Certificate) bool {
	commonLogger.Println("Adding previously removed local certificate")
	if err := cert.TrustCertificates(false, certificates); err == nil {
		commonLogger.Println("Successfully added local certificate")
		return true
	}

	commonLogger.Println("Failed to add local certificate")
	return false
}

func runRevert(args []string) error {
	values, fs := launcherCommonCmd.AdminRevertFlagSet()
	if err := fs.Parse(args); err != nil {
		return err
	}
	internal.SetUp = false
	if values.LogRoot != "" {
		internal.Initialize(values.LogRoot)
	}
	if values.RemoveAll {
		values.UnmapIPs = true
		values.RemoveLocalCert = true
	}
	var removedCertificates []*x509.Certificate
	if values.RemoveLocalCert {
		commonLogger.Println("Removing local certificate")
		var err error
		removedCertificates, err = cert.UntrustCertificates(false)
		if err == nil {
			commonLogger.Println("Successfully removed local certificate")
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				_, ok := <-sigs
				if ok {
					trustCertificates(removedCertificates)
					os.Exit(common.ErrSignal)
				}
			}()
		} else {
			commonLogger.Println("Failed to remove local certificate")
			if !values.RemoveAll {
				os.Exit(internal.ErrLocalCertRemove)
			}
		}
	}
	if values.UnmapIPs {
		commonLogger.Println("Removing IP mappings")
		if err := hosts.RemoveHosts(); err == nil {
			commonLogger.Println("Successfully removed IP mappings")
		} else {
			errorCode := internal.ErrIpMapRemove
			if !values.RemoveAll {
				if removedCertificates != nil {
					if !trustCertificates(removedCertificates) {
						errorCode = internal.ErrIpMapRemoveRevert
					}
				}
			}
			commonLogger.Println("Failed to remove IP mappings")
			if !values.RemoveAll {
				os.Exit(errorCode)
			}
		}
	}
	return nil
}
