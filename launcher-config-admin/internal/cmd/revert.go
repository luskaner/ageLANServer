package cmd

import (
	"crypto/x509"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/pflag"
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
	fs := pflag.NewFlagSet("revert", pflag.ContinueOnError)
	fs.BoolVarP(&launcherCommonCmd.UnmapIPs, "ip", "i", false, "Remove the IP mappings from the local DNS server")
	fs.BoolVarP(&launcherCommonCmd.RemoveLocalCert, "localCert", "l", false, "Remove the certificate from the local machine's trusted root store")
	fs.BoolVarP(&launcherCommonCmd.RemoveAll, "all", "a", false, "Removes all configuration. Equivalent to the rest of the flags being set without fail-fast.")
	commonCmd.LogRootCommand(fs, &logRoot)

	if err := fs.Parse(args); err != nil {
		return err
	}

	internal.SetUp = false
	if logRoot != "" {
		internal.Initialize(logRoot)
	}
	if launcherCommonCmd.RemoveAll {
		launcherCommonCmd.UnmapIPs = true
		launcherCommonCmd.RemoveLocalCert = true
	}
	var removedCertificates []*x509.Certificate
	if launcherCommonCmd.RemoveLocalCert {
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
			if !launcherCommonCmd.RemoveAll {
				os.Exit(internal.ErrLocalCertRemove)
			}
		}
	}
	if launcherCommonCmd.UnmapIPs {
		commonLogger.Println("Removing IP mappings")
		if err := hosts.RemoveHosts(); err == nil {
			commonLogger.Println("Successfully removed IP mappings")
		} else {
			errorCode := internal.ErrIpMapRemove
			if !launcherCommonCmd.RemoveAll {
				if removedCertificates != nil {
					if !trustCertificates(removedCertificates) {
						errorCode = internal.ErrIpMapRemoveRevert
					}
				}
			}
			commonLogger.Println("Failed to remove IP mappings")
			if !launcherCommonCmd.RemoveAll {
				os.Exit(errorCode)
			}
		}
	}
	return nil
}
