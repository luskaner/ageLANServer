package cmd

import (
	"common"
	"config/internal"
	"config/internal/userData"
	"crypto/x509"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	launcherCommon "launcher-common"
	"launcher-common/cmd"
	"launcher-common/executor"
	"os"
	"os/signal"
)

func addUserCert(removedUserCert *x509.Certificate) bool {
	fmt.Println("Adding previously removed user certificate")
	if err := launcherCommon.TrustCertificate(true, removedUserCert); err == nil {
		fmt.Println("Successfully added user certificate")
		return true
	} else {
		fmt.Println("Failed to add user certificate")
		return false
	}
}

func backupMetadata() bool {
	fmt.Println("Backing up previously restored metadata")
	if userData.Metadata.Backup() {
		fmt.Println("Successfully backed up metadata")
		return true
	} else {
		fmt.Println("Failed to back up metadata")
		return false
	}
}

func backupProfiles() bool {
	fmt.Println("Backing up previously restored profiles")
	if userData.BackupProfiles() {
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
	Long:  "Reverts any of the following:\n* Any host mappings to the local DNS server\n* Certificate to the user/local machine's trusted root store\n* User metadata\n* User profiles",
	Run: func(_ *cobra.Command, _ []string) {
		var removedUserCert *x509.Certificate
		var restoredMetadata bool
		var restoredProfiles bool
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, windows.SIGINT, windows.SIGTERM)
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
			cmd.RemoveLocalCert = true
			RemoveUserCert = true
			RestoreMetadata = true
			RestoreProfiles = true
			reverseFailed = false
		}
		if RemoveUserCert {
			fmt.Println("Removing user certificate, accept any dialog that appears...")
			if removedUserCert, _ := launcherCommon.UntrustCertificate(true); removedUserCert != nil {
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
			if userData.Metadata.Restore() {
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
			if userData.RestoreProfiles(reverseFailed) {
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
		if cmd.RemoveLocalCert || cmd.UnmapIPs {
			agentStarted := internal.ConnectAgentIfNeeded() == nil
			if agentStarted {
				fmt.Println("Communicating with config-admin-agent to remove local cert and/or host mappings...")
			} else {
				if isAdmin {
					fmt.Println("Running config-admin to remove local cert and/or host mappings...")
				} else {
					fmt.Println("Running config-admin to remove local cert and/or host mappings, accept any dialog that appears...")
				}
			}
			err, exitCode := internal.RunRevert(cmd.UnmapIPs, cmd.RemoveLocalCert, !cmd.RemoveAll)
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
			fmt.Println("Trying to stop config-admin-agent.")
			err := internal.StopAgentIfNeeded()
			if err == nil {
				if internal.ConnectAgentIfNeededWithRetries(false) {
					fmt.Println("Stopped config-admin-agent")
				} else {
					fmt.Println("Failed to stop config-admin-agent")
					os.Exit(internal.ErrStopAgentVerify)
				}
			} else {
				fmt.Println("Failed to trying stopping config-admin-agent")
				fmt.Println(err)
				os.Exit(internal.ErrStopAgent)
			}
		}
	},
}

var RemoveUserCert bool
var RestoreMetadata bool
var RestoreProfiles bool
var stopAgent bool

func initRevert() {
	cmd.InitRevert(revertCmd)
	revertCmd.Flags().BoolVarP(
		&RemoveUserCert,
		"userCert",
		"u",
		false,
		"Remove the certificate from the user's trusted root store",
	)
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
	rootCmd.AddCommand(revertCmd)
}
