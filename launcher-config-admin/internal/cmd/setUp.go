package cmd

import (
	"crypto/x509"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd"
	launcherCommonHosts "github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal"
	"github.com/luskaner/ageLANServer/launcher-config-admin/internal/hosts"
	"github.com/spf13/pflag"
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
	fs := pflag.NewFlagSet("setup", pflag.ContinueOnError)
	// register flags using launcher-common helpers
	launcherCommonCmd.InitSetUp(fs)
	commonCmd.LogRootCommand(fs, &logRoot)

	if err := fs.Parse(args); err != nil {
		return err
	}

	// decode flags provided by launcher-common helper
	if err := launcherCommonCmd.DecodeSetUpFlags(); err != nil {
		return err
	}

	// validate required flags
	if launcherCommonCmd.GameId == "" {
		return errors.New("required flag 'game' not set")
	}

	// original run body
	internal.SetUp = true
	if logRoot != "" {
		internal.Initialize(logRoot)
	}
	trustedCertificate := false
	if len(launcherCommonCmd.AddLocalCertData) > 0 {
		commonLogger.Println("Adding local certificate")
		crt := cert.BytesToCertificate(launcherCommonCmd.AddLocalCertData)
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
	if len(launcherCommonCmd.MapIP) > 0 {
		commonLogger.Println("Adding IP mappings")
		if ok, _ := launcherCommonHosts.AddHosts(launcherCommonCmd.GameId, "", "", hosts.FlushDns); ok {
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
