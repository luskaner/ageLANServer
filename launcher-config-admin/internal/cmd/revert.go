package cmd

import (
	"crypto/x509"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func trustCertificate(certificate *x509.Certificate) bool {
	fmt.Println("Adding previously removed local certificate")
	if err := cert.TrustCertificate(false, certificate); err == nil {
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
			if removedCertificate, err := cert.UntrustCertificate(false); err == nil {
				fmt.Println("Successfully removed local certificate")
				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
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
		if cmd.UnmapCDN || cmd.UnmapIPs {
			hsts := mapset.NewSet[string]()
			if cmd.UnmapIPs {
				hsts.Add(common.Domain)
			}
			if cmd.UnmapCDN {
				hsts.Add(launcherCommon.CDNDomain)
			}
			fmt.Println("Removing IP mappings")
			if err := hosts.RemoveHosts(hsts); err == nil {
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
