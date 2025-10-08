package cmd

import (
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/cobra"
)

func trustCertificates(certificates []*x509.Certificate) bool {
	fmt.Println("Adding previously removed local certificate")
	if err := cert.TrustCertificates(false, certificates); err == nil {
		fmt.Println("Successfully added local certificate")
		return true
	} else {
		fmt.Println("Failed to add local certificate")
		return false
	}
}

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Reverts configuration",
	Long:  "Removes one or more host mappings from the local DNS server and/or removes a certificate from the local machine's trusted root store",
	Run: func(_ *cobra.Command, _ []string) {
		if cmd.RemoveAll {
			cmd.UnmapIPs = true
			cmd.RemoveLocalCert = true
		}
		var removedCertificates []*x509.Certificate
		if cmd.RemoveLocalCert {
			fmt.Println("Removing local certificate")
			if removedCertificates, err := cert.UntrustCertificates(false); err == nil {
				fmt.Println("Successfully removed local certificate")
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
				fmt.Println("Failed to remove local certificate")
				if !cmd.RemoveAll {
					os.Exit(internal.ErrLocalCertRemove)
				}
			}
		}
		if cmd.UnmapIPs {
			fmt.Println("Removing IP mappings")
			if err := hosts.RemoveHosts(); err == nil {
				fmt.Println("Successfully removed IP mappings")
			} else {
				errorCode := internal.ErrIpMapRemove
				if !cmd.RemoveAll {
					if removedCertificates != nil {
						if !trustCertificates(removedCertificates) {
							errorCode = internal.ErrIpMapRemoveRevert
						}
					}
				}
				fmt.Println("Failed to remove IP mappings")
				if !cmd.RemoveAll {
					os.Exit(errorCode)
				}
			}
		}
	},
}

func initRevert() {
	cmd.InitRevert(revertCmd)
	rootCmd.AddCommand(revertCmd)
}
