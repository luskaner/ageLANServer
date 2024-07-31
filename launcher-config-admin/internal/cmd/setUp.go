package cmd

import (
	"cfgAdmin/internal"
	"common"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	launcherCommon "launcherCommon"
	"launcherCommon/cmd"
	"os"
	"os/signal"
)

func untrustCertificate() bool {
	fmt.Println("Removing previously added local certificate")
	if _, err := launcherCommon.UntrustCertificate(false); err != nil {
		fmt.Println("Successfully removed local certificate")
		return true
	} else {
		fmt.Println("Failed to remove local certificate")
		return false
	}
}

var setUpCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups configuration",
	Long:  "Adds one or more host mappings to the local DNS server and/or adding a certificate to the local machine's trusted root store",
	Run: func(_ *cobra.Command, _ []string) {
		if len(cmd.MapIPs) > 9 {
			fmt.Println("Too many IPs. Up to 9 can be mapped")
			os.Exit(launcherCommon.ErrIpMapAddTooMany)
		}
		trustedCertificate := false
		if len(cmd.AddLocalCertData) > 0 {
			fmt.Println("Adding local certificate")
			cert := launcherCommon.BytesToCertificate(cmd.AddLocalCertData)
			if cert == nil {
				fmt.Println("Failed to parse certificate")
				os.Exit(internal.ErrLocalCertAddParse)
			}
			if err := launcherCommon.TrustCertificate(false, cert); err == nil {
				fmt.Println("Successfully added local certificate")
				trustedCertificate = true
				sigs := make(chan os.Signal, 1)
				signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
				go func() {
					_, ok := <-sigs
					if ok {
						untrustCertificate()
						os.Exit(common.ErrSignal)
					}
				}()
			} else {
				fmt.Println("Failed to add local certificate")
				os.Exit(internal.ErrLocalCertAdd)
			}
		}
		if len(cmd.MapIPs) > 0 {
			fmt.Println("Adding IP mappings")
			ipStrSet := mapset.NewSet[string]()
			for _, ip := range cmd.MapIPs {
				ipStrSet.Add(ip.String())
			}
			if ok, _ := internal.AddHosts(ipStrSet); ok {
				fmt.Println("Successfully added IP mappings")
			} else {
				errorCode := internal.ErrIpMapAdd
				if trustedCertificate {
					if !untrustCertificate() {
						errorCode = internal.ErrIpMapAddRevert
					}
				}
				fmt.Println("Failed to add IP mappings")
				os.Exit(errorCode)
			}
		}
	},
}

func initSetUp() {
	cmd.InitSetUp(setUpCmd)
	rootCmd.AddCommand(setUpCmd)
}
