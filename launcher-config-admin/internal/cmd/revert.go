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
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/cobra"
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

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Reverts configuration",
	Long:  "Removes one or more host mappings from the local DNS server and/or removes a certificate from the local machine's trusted root store",
	Run: func(_ *cobra.Command, _ []string) {
		internal.SetUp = false
		if logRoot != "" {
			internal.Initialize(logRoot)
		}
		if cmd.RemoveAll {
			cmd.UnmapIPs = true
			cmd.RemoveLocalCert = true
		}
		var removedCertificates []*x509.Certificate
		if cmd.RemoveLocalCert {
			commonLogger.Println("Removing local certificate")
			if removedCertificates, err := cert.UntrustCertificates(false); err == nil {
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
				if !cmd.RemoveAll {
					os.Exit(internal.ErrLocalCertRemove)
				}
			}
		}
		if cmd.UnmapIPs {
			commonLogger.Println("Removing IP mappings")
			if err := hosts.RemoveHosts(); err == nil {
				commonLogger.Println("Successfully removed IP mappings")
			} else {
				errorCode := internal.ErrIpMapRemove
				if !cmd.RemoveAll {
					if removedCertificates != nil {
						if !trustCertificates(removedCertificates) {
							errorCode = internal.ErrIpMapRemoveRevert
						}
					}
				}
				commonLogger.Println("Failed to remove IP mappings")
				if !cmd.RemoveAll {
					os.Exit(errorCode)
				}
			}
		}
	},
}

func initRevert() {
	cmd.InitRevert(revertCmd)
	commonCmd.LogRootCommand(revertCmd.Flags(), &logRoot)
	rootCmd.AddCommand(revertCmd)
}
