package cmd

import (
	"crypto/x509"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
	"github.com/luskaner/ageLANServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/ageLANServer/launcher-config/internal/userData"
	"github.com/spf13/cobra"
)

func addUserCerts(removedUserCerts []*x509.Certificate) bool {
	fmt.Println("Adding previously removed user certificate")
	if err := wrapper.AddUserCerts(removedUserCerts); err == nil {
		fmt.Println("Successfully added user certificate")
		return true
	} else {
		fmt.Println("Failed to add user certificate")
		return false
	}
}

func backupMetadata() bool {
	fmt.Println("Backing up previously restored metadata")
	if userData.Metadata(cmd.GameId).Backup(gameId) {
		fmt.Println("Successfully backed up metadata")
		return true
	} else {
		fmt.Println("Failed to back up metadata")
		return false
	}
}

func backupProfiles() bool {
	fmt.Println("Backing up previously restored profiles")
	if userData.BackupProfiles(cmd.GameId) {
		fmt.Println("Successfully backed up profiles")
		return true
	} else {
		fmt.Println("Failed to back up profiles")
		return false
	}
}

func addCaCerts(removedCaCerts []*x509.Certificate) bool {
	fmt.Println("Restoring previously added game's certificate store...")
	if err := internal.NewCACert(cmd.GameId, gamePath).Append(removedCaCerts); err == nil {
		fmt.Println("Successfully restored game's certificate store.")
		return true
	} else {
		fmt.Println("Failed to restore game's certificate store.")
		return false
	}
}

func undoRevert() {
	if !cmd.RemoveAll {
		if removedCaCerts != nil {
			addCaCerts(removedCaCerts)
		}
		if removedUserCerts != nil {
			addUserCerts(removedUserCerts)
		}
		if restoredMetadata {
			backupMetadata()
		}
		if restoredProfiles {
			backupProfiles()
		}
		os.Exit(errorCode)
	}
}

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Reverts configuration",
	Long:  "Reverts any of the following:\n* Any host mappings to the local DNS server\n* Certificate to the " + storeString + " machine's trusted root store\n* User metadata\n* User profiles",
	Run: func(_ *cobra.Command, _ []string) {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			_, ok := <-sigs
			if ok {
				undoRevert()
				os.Exit(common.ErrSignal)
			}
		}()
		isAdmin := executor.IsAdmin()
		reverseFailed := true
		if cmd.RemoveAll {
			cmd.UnmapIPs = true
			cmd.RemoveLocalCert = true
			if runtime.GOOS != "linux" {
				doRemoveUserCert = true
			}
			doRestoreMetadata = true
			doRestoreProfiles = true
			doRestoreCaStoreCert = true
			reverseFailed = false
		}
		if cmd.GameId == common.GameAoE1 {
			doRestoreMetadata = false
			doRestoreCaStoreCert = false
		}
		if (restoredMetadata || restoredProfiles) && !common.SupportedGames.ContainsOne(cmd.GameId) {
			fmt.Println("Invalid game type")
			errorCode = launcherCommon.ErrInvalidGame
			undoRevert()
		}
		fmt.Printf("Reverting configuration for %s...\n", cmd.GameId)
		if doRemoveUserCert {
			fmt.Println("Removing user certificates, authorize it if needed...")
			if removedUserCerts, _ := wrapper.RemoveUserCerts(); removedUserCerts != nil {
				fmt.Println("Successfully removed user certificates")
			} else {
				fmt.Println("Failed to remove user certificates")
				errorCode = internal.ErrUserCertRemove
				undoRevert()
			}
		}
		if doRestoreMetadata {
			fmt.Println("Restoring metadata")
			if userData.Metadata(cmd.GameId).Restore(cmd.GameId) {
				fmt.Println("Successfully restored metadata")
				restoredMetadata = true
			} else {
				fmt.Println("Failed to restore metadata")
				errorCode = internal.ErrMetadataRestore
				undoRevert()
			}
		}
		if doRestoreProfiles {
			fmt.Println("Restoring profiles")
			if userData.RestoreProfiles(cmd.GameId, reverseFailed) {
				fmt.Println("Successfully restored profiles")
				restoredProfiles = true
			} else {
				fmt.Println("Failed to restore profiles")
				errorCode = internal.ErrProfilesRestore
				undoRevert()
			}
		}
		if doRestoreCaStoreCert {
			fmt.Println("Restoring original certificate game's store...")
			if gamePath == "" {
				fmt.Println("Game path is required to restore the original game's store")
				errorCode = internal.ErrGamePathMissing
				undoRevert()
			}
			cert := internal.NewCACert(cmd.GameId, gamePath)
			var err error
			if err, removedCaCerts = cert.Restore(); err == nil {
				fmt.Println("Successfully restored original game's store.")
			} else {
				fmt.Println("Failed to restore original game's store.")
				fmt.Println("Received error:")
				fmt.Println(err)
				errorCode = internal.ErrGameCertRestore
				undoRevert()
			}
		}
		var agentConnected bool
		if cmd.RemoveLocalCert || cmd.UnmapIPs {
			agentConnected = internal.ConnectAgentIfNeeded() == nil
			if agentConnected {
				fmt.Println("Communicating with 'config-admin-agent' to remove local cert and/or host mappings...")
			} else {
				fmt.Print("Running 'config-admin' to remove local cert and/or host mappings")
				if !isAdmin {
					fmt.Print(", authorize it if needed")
				}
				fmt.Println("...")
			}
			var err error
			err, errorCode = internal.RunRevert(cmd.UnmapIPs, cmd.RemoveLocalCert, !cmd.RemoveAll)
			if err == nil && errorCode == common.ErrSuccess {
				if agentConnected {
					fmt.Println("Successfully communicated with 'config-admin-agent'")
				} else {
					fmt.Println("Successfully ran 'config-admin'")
				}
			} else {
				if err != nil {
					fmt.Println("Received error:")
					fmt.Println(err)
				}
				if errorCode != common.ErrSuccess {
					fmt.Println("Received exit code:")
					fmt.Println(errorCode)
				}
				errorCode = internal.ErrAdminRevert
				undoRevert()
				if agentConnected {
					fmt.Println("Failed to communicate with 'config-admin-agent'")
				} else {
					fmt.Println("Failed to run 'config-admin'")
				}
			}
		}
		// Ignore previous error if we don't failfast
		if cmd.RemoveAll {
			errorCode = common.ErrSuccess
		}
		if errorCode == common.ErrSuccess && hostFilePath != "" {
			_ = os.Remove(hostFilePath)
		}
		if errorCode == common.ErrSuccess && certFilePath != "" {
			_ = os.Remove(certFilePath)
		}
		if stopAgent {
			failedStopAgent := true
			if agentConnected {
				fmt.Println("Trying to stop 'config-admin-agent'.")
				err := internal.StopAgentIfNeeded()
				if err == nil {
					if internal.ConnectAgentIfNeededWithRetries(false) {
						fmt.Println("Stopped 'config-admin-agent'")
						failedStopAgent = false
					} else {
						fmt.Println("Failed to stop 'config-admin-agent'")
					}
				} else {
					fmt.Println("Failed to trying stopping 'config-admin-agent'")
					fmt.Println(err)
				}
			}
			if failedStopAgent {
				exeFileName := common.GetExeFileName(true, common.LauncherConfigAdminAgent)
				if pid, proc, err := commonProcess.Process(exeFileName); err == nil {
					if isAdmin {
						if err := commonProcess.KillPidProc(pid, proc); err == nil {
							fmt.Println("Successfully killed 'config-admin-agent'.")
							failedStopAgent = false
						} else {
							fmt.Println("Failed to kill 'config-admin-agent'")
						}
					} else {
						fmt.Println("Re-run as administrator to kill 'config-admin-agent'")
					}
				} else {
					failedStopAgent = false
				}
			}
			if failedStopAgent && errorCode == common.ErrSuccess {
				errorCode = internal.ErrRevertStopAgent
			}
			os.Exit(errorCode)
		}
	},
}

var doRemoveUserCert bool
var doRestoreMetadata bool
var doRestoreProfiles bool
var doRestoreCaStoreCert bool
var stopAgent bool
var gameId string

// State
var removedUserCerts []*x509.Certificate
var removedCaCerts []*x509.Certificate
var restoredMetadata bool
var restoredProfiles bool

func InitRevert() {
	if runtime.GOOS != "linux" {
		storeString = "user/" + storeString
	}
	cmd.InitRevert(revertCmd)
	addGamePathFlags(revertCmd)
	commonCmd.GameVarCommand(revertCmd.Flags(), &cmd.GameId)
	err := revertCmd.MarkFlagRequired("game")
	if err != nil {
		panic(err)
	}
	revertCmd.Flags().StringVarP(
		&hostFilePath,
		"hostFilePath",
		"o",
		"",
		"Path to the host file.",
	)
	revertCmd.Flags().StringVarP(
		&certFilePath,
		"certFilePath",
		"t",
		"",
		"Path to the certificate file.",
	)
	if runtime.GOOS != "linux" {
		revertCmd.Flags().BoolVarP(
			&doRemoveUserCert,
			"userCert",
			"u",
			false,
			"Remove the certificate from the user's trusted root store",
		)
	}
	revertCmd.Flags().BoolVarP(
		&doRestoreMetadata,
		"metadata",
		"m",
		false,
		"Restore metadata. Not compatible with AoE:DE",
	)
	revertCmd.Flags().BoolVarP(
		&doRestoreProfiles,
		"profiles",
		"p",
		false,
		"Restore profiles",
	)
	revertCmd.Flags().BoolVarP(
		&doRestoreCaStoreCert,
		"caStoreCert",
		"s",
		false,
		"Restore the game's trusted root store. For all except AoE I: DE.",
	)
	revertCmd.Flags().BoolVarP(
		&stopAgent,
		"stopAgent",
		"g",
		false,
		"Stop the 'config-admin-agent' if it is running after all operations",
	)
	err = revertCmd.Flags().MarkHidden("stopAgent")
	if err != nil {
		panic(err)
	}
	RootCmd.AddCommand(revertCmd)
}
