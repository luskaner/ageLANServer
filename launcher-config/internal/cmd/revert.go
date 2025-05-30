package cmd

import (
	"crypto/x509"
	"fmt"
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
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
	if userData.Metadata(gameId).Backup(gameId) {
		fmt.Println("Successfully backed up metadata")
		return true
	} else {
		fmt.Println("Failed to back up metadata")
		return false
	}
}

func backupProfiles() bool {
	fmt.Println("Backing up previously restored profiles")
	if userData.BackupProfiles(gameId) {
		fmt.Println("Successfully backed up profiles")
		return true
	} else {
		fmt.Println("Failed to back up profiles")
		return false
	}
}

func undoRevert(removedUserCerts []*x509.Certificate, restoredMetadata bool, restoredProfiles bool) {
	if removedUserCerts != nil {
		addUserCerts(removedUserCerts)
	}
	if restoredMetadata {
		backupMetadata()
	}
	if restoredProfiles {
		backupProfiles()
	}
}

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Reverts configuration",
	Long:  "Reverts any of the following:\n* Any host mappings to the local DNS server\n* Certificate to the " + storeString + " machine's trusted root store\n* User metadata\n* User profiles",
	Run: func(_ *cobra.Command, _ []string) {
		var removedUserCerts []*x509.Certificate
		var restoredMetadata bool
		var restoredProfiles bool
		var errorCode int
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			_, ok := <-sigs
			if ok {
				undoRevert(removedUserCerts, restoredMetadata, restoredProfiles)
				os.Exit(common.ErrSignal)
			}
		}()
		isAdmin := executor.IsAdmin()
		reverseFailed := true
		if cmd.RemoveAll {
			cmd.UnmapIPs = true
			cmd.UnmapCDN = true
			cmd.RemoveLocalCert = true
			if runtime.GOOS != "linux" {
				RemoveUserCert = true
			}
			RestoreMetadata = true
			RestoreProfiles = true
			reverseFailed = false
		}
		if gameId == common.GameAoE1 {
			RestoreMetadata = false
		}
		if (restoredMetadata || restoredProfiles) && !common.SupportedGames.ContainsOne(gameId) {
			fmt.Println("Invalid game type")
			os.Exit(launcherCommon.ErrInvalidGame)
		}
		fmt.Printf("Reverting configuration for %s...\n", gameId)
		if RemoveUserCert {
			fmt.Println("Removing user certificates, authorize it if needed...")
			if removedUserCerts, _ := wrapper.RemoveUserCerts(); removedUserCerts != nil {
				fmt.Println("Successfully removed user certificates")
			} else {
				fmt.Println("Failed to remove user certificates")
				if !cmd.RemoveAll {
					os.Exit(internal.ErrUserCertRemove)
				}
			}
		}
		if RestoreMetadata {
			fmt.Println("Restoring metadata")
			if userData.Metadata(gameId).Restore(gameId) {
				fmt.Println("Successfully restored metadata")
				restoredMetadata = true
			} else {
				errorCode = internal.ErrMetadataRestore
				if !cmd.RemoveAll {
					if removedUserCerts != nil {
						if !addUserCerts(removedUserCerts) {
							errorCode = internal.ErrMetadataRestoreRevert
						}
					}
				}
				fmt.Println("Failed to restore metadata")
				if !cmd.RemoveAll {
					os.Exit(errorCode)
				}
			}
		}
		if RestoreProfiles {
			fmt.Println("Restoring profiles")
			if userData.RestoreProfiles(gameId, reverseFailed) {
				fmt.Println("Successfully restored profiles")
				restoredProfiles = true
			} else {
				errorCode := internal.ErrProfilesRestore
				if !cmd.RemoveAll {
					if removedUserCerts != nil {
						if !addUserCerts(removedUserCerts) {
							errorCode = internal.ErrProfilesRestoreRevert
						}
					}
					if restoredMetadata {
						if !backupMetadata() {
							errorCode = internal.ErrProfilesRestoreRevert
						}
					}
				}
				fmt.Println("Failed to restore profiles")
				if !cmd.RemoveAll {
					os.Exit(errorCode)
				}
			}
		}
		var agentConnected bool
		if cmd.RemoveLocalCert || cmd.UnmapIPs || cmd.UnmapCDN {
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
			err, errorCode = internal.RunRevert(cmd.UnmapIPs, cmd.RemoveLocalCert, cmd.UnmapCDN, !cmd.RemoveAll)
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
				if !cmd.RemoveAll {
					if removedUserCerts != nil {
						if !addUserCerts(removedUserCerts) {
							errorCode = internal.ErrAdminRevertRevert
						}
					}
					if restoredMetadata {
						if !backupMetadata() {
							errorCode = internal.ErrAdminRevertRevert
						}
					}
					if restoredProfiles {
						if !backupProfiles() {
							errorCode = internal.ErrAdminRevertRevert
						}
					}
				}
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
						if err := commonProcess.KillProc(pid, proc); err == nil {
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

var RemoveUserCert bool
var RestoreMetadata bool
var RestoreProfiles bool
var stopAgent bool

func InitRevert() {
	if runtime.GOOS != "linux" {
		storeString = "user/" + storeString
	}
	cmd.InitRevert(revertCmd)
	commonCmd.GameVarCommand(revertCmd.Flags(), &gameId)
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
			&RemoveUserCert,
			"userCert",
			"u",
			false,
			"Remove the certificate from the user's trusted root store",
		)
	}
	revertCmd.Flags().BoolVarP(
		&RestoreMetadata,
		"metadata",
		"m",
		false,
		"Restore metadata. Not compatible with AoE:DE",
	)
	revertCmd.Flags().BoolVarP(
		&RestoreProfiles,
		"profiles",
		"p",
		false,
		"Restore profiles",
	)
	revertCmd.Flags().BoolVarP(
		&stopAgent,
		"stopAgent",
		"g",
		false,
		"Stop the 'config-admin-agent' if it is running after all operations",
	)
	err := revertCmd.Flags().MarkHidden("stopAgent")
	if err != nil {
		panic(err)
	}
	RootCmd.AddCommand(revertCmd)
}
