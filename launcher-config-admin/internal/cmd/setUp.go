package cmd

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/cert"
	"github.com/luskaner/aoe2DELanServer/launcher-common/cmd"
	"github.com/luskaner/aoe2DELanServer/launcher-config-admin/internal"
	"github.com/luskaner/aoe2DELanServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func untrustCertificate() bool {
	fmt.Println("Removing previously added local certificate")
	if _, err := cert.UntrustCertificate(false); err != nil {
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
			crt := cert.BytesToCertificate(cmd.AddLocalCertData)
			if crt == nil {
				fmt.Println("Failed to parse certificate")
				os.Exit(internal.ErrLocalCertAddParse)
			}
			if err := cert.TrustCertificate(false, crt); err == nil {
				fmt.Println("Successfully added local certificate")
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
				fmt.Println("Failed to add local certificate")
				os.Exit(internal.ErrLocalCertAdd)
			}
		}
		if len(cmd.MapIPs) > 0 || cmd.MapCDN {
			fmt.Println("Adding IP mappings")
			mappings := make(map[string]mapset.Set[string])
			if len(cmd.MapIPs) > 0 {
				mappings[common.Domain] = mapset.NewSet[string]()
				for _, ip := range cmd.MapIPs {
					mappings[common.Domain].Add(ip.String())
				}
			}
			if cmd.MapCDN {
				mappings[launcherCommon.CDNDomain] = mapset.NewSet[string]()
				mappings[launcherCommon.CDNDomain].Add(launcherCommon.CDNIP)
			}
			if ok, _ := hosts.AddHosts(mappings); ok {
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
