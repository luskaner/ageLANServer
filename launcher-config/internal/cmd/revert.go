package cmd

import (
	"crypto/x509"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	commonCmd "github.com/luskaner/aoe2DELanServer/common/cmd"
	"github.com/luskaner/aoe2DELanServer/common/executor"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/cmd"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal/userData"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func addUserCert(removedUserCert *x509.Certificate) bool {
	fmt.Println("Adding previously removed user certificate")
	if err := wrapper.AddUserCert(removedUserCert); err == nil {
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

func undoRevert(removedUserCert *x509.Certificate, restoredMetadata bool, restoredProfiles bool) {
	if removedUserCert != nil {
		addUserCert(removedUserCert)
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
		var removedUserCert *x509.Certificate
		var restoredMetadata bool
		var restoredProfiles bool
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			_, ok := <-sigs
			if ok {
				undoRevert(removedUserCert, restoredMetadata, restoredProfiles)
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
		if (restoredMetadata || restoredProfiles) && !common.SupportedGames.ContainsOne(gameId) {
			fmt.Println("Invalid game type")
			os.Exit(launcherCommon.ErrInvalidGame)
		}
		if RemoveUserCert {
			fmt.Println("Removing user certificate, authorize it if needed...")
			if removedUserCert, _ := wrapper.RemoveUserCert(); removedUserCert != nil {
				fmt.Println("Successfully removed user certificate")
			} else {
				fmt.Println("Failed to remove user certificate")
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
				errorCode := internal.ErrMetadataRestore
				if !cmd.RemoveAll {
					if removedUserCert != nil {
						if !addUserCert(removedUserCert) {
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
					if removedUserCert != nil {
						if !addUserCert(removedUserCert) {
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
		if cmd.RemoveLocalCert || cmd.UnmapIPs || cmd.UnmapCDN {
			agentStarted := internal.ConnectAgentIfNeeded() == nil
			if agentStarted {
				fmt.Println("Communicating with config-admin-agent to remove local cert and/or host mappings...")
			} else {
				if isAdmin {
					fmt.Println("Running config-admin to remove local cert and/or host mappings...")
				} else {
					fmt.Println("Running config-admin to remove local cert and/or host mappings, authorize it if needed...")
				}
			}
			err, exitCode := internal.RunRevert(cmd.UnmapIPs, cmd.RemoveLocalCert, cmd.UnmapCDN, !cmd.RemoveAll)
			if err == nil && exitCode == common.ErrSuccess {
				if agentStarted {
					fmt.Println("Successfully communicated with config-admin-agent")
				} else {
					fmt.Println("Successfully ran config-admin")
				}
			} else {
				if err != nil {
					fmt.Println("Received error:")
					fmt.Println(err)
				}
				if exitCode != common.ErrSuccess {
					fmt.Println("Received exit code:")
					fmt.Println(exitCode)
				}
				errorCode := internal.ErrAdminRevert
				if !cmd.RemoveAll {
					if removedUserCert != nil {
						if !addUserCert(removedUserCert) {
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
				if agentStarted {
					fmt.Println("Failed to communicate with config-admin-agent")
				} else {
					fmt.Println("Failed to run config-admin")
				}
				os.Exit(errorCode)
			}
		}
		if stopAgent && internal.ConnectAgentIfNeeded() == nil {
			errorCode := common.ErrSuccess
			fmt.Println("Trying to stop config-admin-agent.")
			err := internal.StopAgentIfNeeded()
			failedStopAgent := true
			if err == nil {
				if internal.ConnectAgentIfNeededWithRetries(false) {
					fmt.Println("Stopped config-admin-agent")
					failedStopAgent = false
				} else {
					fmt.Println("Failed to stop config-admin-agent")
				}
			} else {
				fmt.Println("Failed to trying stopping config-admin-agent")
				fmt.Println(err)
			}
			if isAdmin && failedStopAgent {
				_, err := commonProcess.Kill(common.GetExeFileName(true, common.LauncherConfigAdminAgent))
				if err == nil {
					fmt.Println("Successfully killed config-admin-agent.")
				}
			}
			os.Exit(errorCode)
		}
	},
}

var RemoveUserCert bool
var RestoreMetadata bool
var RestoreProfiles bool
var stopAgent bool
var gameId string

func InitRevert() {
	if runtime.GOOS != "linux" {
		storeString = "user/" + storeString
	}
	cmd.InitRevert(revertCmd)
	commonCmd.GameVarCommand(revertCmd.Flags(), &gameId)
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
		"Restore metadata",
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
		"Stop the config-admin-agent if it is running after all operations are successful",
	)
	err := revertCmd.Flags().MarkHidden("stopAgent")
	if err != nil {
		panic(err)
	}
	RootCmd.AddCommand(revertCmd)
}
