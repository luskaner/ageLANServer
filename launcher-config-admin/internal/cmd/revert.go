package cmd

import (
	"admin/internal"
	"common"
	"crypto/x509"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	launcherCommon "launcher-common"
	"launcher-common/cmd"
	"os"
	"os/signal"
)

func trustCertificate(certificate *x509.Certificate) bool {
	fmt.Println("Adding previously removed local certificate")
	if err := launcherCommon.TrustCertificate(false, certificate); err == nil {
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
		var removedCertificate *x509.Certificate
		if cmd.RemoveLocalCert {
			fmt.Println("Removing local certificate")
			if removedCertificate, err := launcherCommon.UntrustCertificate(false); err == nil {
				fmt.Println("Successfully removed local certificate")
				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
				go func() {
					_, ok := <-sigs
					if ok {
						trustCertificate(removedCertificate)
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
			if err := internal.RemoveHosts(); err == nil {
				fmt.Println("Successfully removed IP mappings")
			} else {
				errorCode := internal.ErrIpMapRemove
				if !cmd.RemoveAll {
					if removedCertificate != nil {
						if !trustCertificate(removedCertificate) {
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
