package cmd

import (
	"crypto/x509"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	launcherCommonHosts "github.com/luskaner/ageLANServer/common/hosts"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
)

func untrustCertificate() bool {
	commonLogger.Println("Removing previously added local certificate")
	if _, err := cert.UntrustCertificates(false); err == nil {
		commonLogger.Println("Successfully removed local certificate")
		return true
	}

	commonLogger.Println("Failed to remove local certificate")
	return false
}

func runSetUp(args []string) error {
	values, fs := launcherCommonCmd.AdminSetupFlagSet()
	if err := fs.Parse(args); err != nil {
		return err
	}

	// validate required flags
	if values.GameId == "" {
		return errors.New("required flag 'game' not set")
	}

	internal.SetUp = true
	if values.LogRoot != "" {
		internal.Initialize(values.LogRoot)
	}
	trustedCertificate := false
	if len(values.AddLocalCertData) > 0 {
		commonLogger.Println("Adding local certificate")
		crt := common.BytesToCertificate(values.AddLocalCertData)
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
	if len(values.MapIp) > 0 {
		commonLogger.Println("Adding IP mappings")
		if ok, _ := launcherCommonHosts.AddHosts(values.MapIp, values.GameId, "", "", hosts.FlushDns); ok {
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
	return nil
}
