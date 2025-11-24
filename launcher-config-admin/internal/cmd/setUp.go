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
	launcherCommonHosts "github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/cobra"
)

func untrustCertificate() bool {
	commonLogger.Println("Removing previously added local certificate")
	if _, err := cert.UntrustCertificates(false); err == nil {
		commonLogger.Println("Successfully removed local certificate")
		return true
	} else {
		commonLogger.Println("Failed to remove local certificate")
		return false
	}
}

var setUpCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups configuration",
	Long:  "Adds one or more host mappings to the local DNS server and/or adding a certificate to the local machine's trusted root store",
	Run: func(_ *cobra.Command, _ []string) {
		internal.SetUp = true
		if logRoot != "" {
			internal.Initialize(logRoot)
		}
		trustedCertificate := false
		if len(cmd.AddLocalCertData) > 0 {
			commonLogger.Println("Adding local certificate")
			crt := cert.BytesToCertificate(cmd.AddLocalCertData)
			if crt == nil {
				commonLogger.Println("Failed to parse certificate")
				os.Exit(internal.ErrLocalCertAddParse)
			}
			if err := cert.TrustCertificates(false, []*x509.Certificate{crt}); err == nil {
				commonLogger.Println("Successfully added local certificate")
				trustedCertificate = true
				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
				go func() {
					_, ok := <-sigs
					if ok {
						untrustCertificate()
						os.Exit(common.ErrSignal)
					}
				}()
			} else {
				commonLogger.Println("Failed to add local certificate")
				os.Exit(internal.ErrLocalCertAdd)
			}
		}
		if len(cmd.MapIP) > 0 {
			commonLogger.Println("Adding IP mappings")
			if ok, _ := launcherCommonHosts.AddHosts(cmd.GameId, "", "", hosts.FlushDns); ok {
				commonLogger.Println("Successfully added IP mappings")
			} else {
				errorCode := internal.ErrIpMapAdd
				if trustedCertificate {
					if !untrustCertificate() {
						errorCode = internal.ErrIpMapAddRevert
					}
				}
				commonLogger.Println("Failed to add IP mappings")
				os.Exit(errorCode)
			}
		}
	},
}

func initSetUp() {
	cmd.InitSetUp(setUpCmd)
	commonCmd.LogRootCommand(setUpCmd.Flags(), &logRoot)
	rootCmd.AddCommand(setUpCmd)
}
