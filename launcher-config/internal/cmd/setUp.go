package cmd

import (
	"crypto/x509"
	"fmt"
	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executor"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
	"github.com/luskaner/ageLANServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/ageLANServer/launcher-config/internal/userData"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func removeUserCert() bool {
	fmt.Println("Removing previously added user certificate, authorize it if needed ...")
	if _, err := wrapper.RemoveUserCerts(); err == nil {
		fmt.Println("Successfully removed user certificate")
		return true
	} else {
		fmt.Println("Failed to remove user certificate")
		return false
	}
}

func restoreMetadata() bool {
	fmt.Println("Restoring previously backed up metadata")
	if userData.Metadata(gameId).Restore(windowsUserProfilePath, gameId) {
		fmt.Println("Successfully restored metadata")
		return true
	} else {
		fmt.Println("Failed to restore metadata")
		return false
	}
}

func restoreProfiles() bool {
	fmt.Println("Restoring previously backed up profiles")
	if userData.RestoreProfiles(windowsUserProfilePath, gameId, true) {
		fmt.Println("Successfully restored profiles")
		return true
	} else {
		fmt.Println("Failed to restore profiles")
		return false
	}
}

func undoSetUp(addedUserCert bool, backedUpMetadata bool, backedUpProfiles bool) {
	if addedUserCert {
		removeUserCert()
	}
	if backedUpMetadata {
		restoreMetadata()
	}
	if backedUpProfiles {
		restoreProfiles()
	}
}

var AddUserCertData []byte
var BackupMetadata bool
var BackupProfiles bool
var agentStart bool
var agentEndOnError bool
var windowsUserProfilePath string
var storeString = "local"

var setUpCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups configuration",
	Long:  "Adds any of the following:\n* One or more host mappings to the local DNS resolver\n* Certificate to the " + storeString + " machine's trusted root store\n* Backup user metadata\n* Backup user profiles",
	Run: func(_ *cobra.Command, _ []string) {
		var addedUserCert bool
		var backedUpMetadata bool
		var backedUpProfiles bool
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			_, ok := <-sigs
			if ok {
				undoSetUp(addedUserCert, backedUpMetadata, backedUpProfiles)
				os.Exit(common.ErrSignal)
			}
		}()
		if gameId == common.GameAoE1 {
			BackupMetadata = false
		}
		if (BackupMetadata || BackupProfiles) && !common.SupportedGames.ContainsOne(gameId) {
			fmt.Println("Invalid game type")
			os.Exit(launcherCommon.ErrInvalidGame)
		}
		var addLocalCertData []byte = nil
		if certFilePath != "" {
			if len(cmd.AddLocalCertData) == 0 {
				fmt.Println("Certificate file path is set but no local certificate data is provided")
				os.Exit(internal.ErrMissingLocalCertData)
			}
		} else {
			addLocalCertData = cmd.AddLocalCertData
		}
		fmt.Printf("Setting up configuration for %s...\n", gameId)
		isAdmin := executor.IsAdmin()
		if AddUserCertData != nil {
			fmt.Println("Adding user certificate, authorize it if needed...")
			crt := wrapper.BytesToCertificate(AddUserCertData)
			if crt == nil {
				fmt.Println("Failed to parse certificate")
				os.Exit(internal.ErrUserCertAddParse)
			}
			if err := wrapper.AddUserCerts([]*x509.Certificate{crt}); err == nil {
				fmt.Println("Successfully added user certificate")
				addedUserCert = true
			} else {
				fmt.Println("Failed to add user certificate")
				fmt.Println("Error message: " + err.Error())
				os.Exit(internal.ErrUserCertAdd)
			}
		}
		if BackupMetadata {
			fmt.Println("Backing up metadata")
			if userData.Metadata(gameId).Backup(windowsUserProfilePath, gameId) {
				fmt.Println("Successfully backed up metadata")
				backedUpMetadata = true
			} else {
				errorCode := internal.ErrMetadataBackup
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrMetadataBackupRevert
					}
				}
				fmt.Println("Failed to back up metadata")
				os.Exit(errorCode)
			}
		}
		if BackupProfiles {
			fmt.Println("Backing up profiles")
			if userData.BackupProfiles(windowsUserProfilePath, gameId) {
				fmt.Println("Successfully backed up profiles")
				backedUpProfiles = true
			} else {
				errorCode := internal.ErrProfilesBackup
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrProfilesBackupRevert
					}
				}
				if backedUpMetadata {
					if !restoreMetadata() {
						errorCode = internal.ErrProfilesBackupRevert
					}
				}
				fmt.Println("Failed to back up profiles")
				os.Exit(errorCode)
			}
		}
		var hostToMap string
		if hostFilePath == "" {
			hostToMap = cmd.MapIP.String()
		} else {
			if cmd.MapCDN || cmd.MapIP != nil {
				if ok, _ := hosts.AddHosts(hostFilePath, hosts.WindowsLineEnding, nil); ok {
					fmt.Println("Successfully added host mappings")
				} else {
					fmt.Println("Failed to add host mappings")
					_ = os.Remove(hostFilePath)
					errorCode := internal.ErrHostsAdd
					if addedUserCert {
						if !removeUserCert() {
							errorCode = internal.ErrAdminSetupRevert
						}
					}
					if backedUpMetadata {
						if !restoreMetadata() {
							errorCode = internal.ErrAdminSetupRevert
						}
					}
					if backedUpProfiles {
						if !restoreProfiles() {
							errorCode = internal.ErrAdminSetupRevert
						}
					}
					os.Exit(errorCode)
				}
			}
			cmd.MapCDN = false
		}
		if certFilePath != "" {
			certFile, err := os.Create(certFilePath)
			if err == nil {
				err = cert.WriteAsPem(cmd.AddLocalCertData, certFile)
				if err != nil {
					_ = certFile.Close()
				}
			}
			if err != nil {
				fmt.Println("Error saving certificate file:", err)
				_ = os.Remove(hostFilePath)
				_ = os.Remove(certFilePath)
				errorCode := internal.ErrUserCertAdd
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if backedUpMetadata {
					if !restoreMetadata() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if backedUpProfiles {
					if !restoreProfiles() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				os.Exit(errorCode)
			}
		}
		if addLocalCertData != nil || hostToMap != "" || cmd.MapCDN {
			agentStarted := internal.ConnectAgentIfNeeded() == nil
			if !agentStarted && agentStart && !isAdmin {
				result := internal.StartAgentIfNeeded()
				if !result.Success() {
					fmt.Println("Failed to start 'config-admin-agent'")
					if result.Err != nil {
						fmt.Println(result.Err)
					}
					if result.ExitCode != common.ErrSuccess {
						fmt.Println(result.ExitCode)
					}
					os.Exit(internal.ErrStartAgent)
				} else {
					agentStarted = internal.ConnectAgentIfNeededWithRetries(true)
					if !agentStarted {
						fmt.Println("Failed to connect to 'config-admin-agent' after starting it. Kill it using the task manager.")
						os.Exit(internal.ErrStartAgentVerify)
					}
				}
			}
			if agentStarted {
				fmt.Println("Communicating with 'config-admin-agent' to add local cert and/or host mappings...")
			} else {
				fmt.Print("Running 'config-admin' to add local cert and/or host mappings")
				if !isAdmin {
					fmt.Print(", authorize it if needed")
				}
				fmt.Println("...")
			}
			err, exitCode := internal.RunSetUp(hostToMap, addLocalCertData, cmd.MapCDN)
			if err == nil && exitCode == common.ErrSuccess {
				if agentStarted {
					fmt.Println("Successfully communicated with 'config-admin-agent'")
				} else {
					fmt.Println("Successfully ran 'config-admin'")
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
				errorCode := internal.ErrAdminSetup
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if backedUpMetadata {
					if !restoreMetadata() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if backedUpProfiles {
					if !restoreProfiles() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if hostFilePath != "" {
					_ = os.Remove(hostFilePath)
				}
				if certFilePath != "" {
					_ = os.Remove(certFilePath)
				}
				if agentStarted {
					fmt.Println("Failed to communicate with 'config-admin-agent'. Communicating with it to shutdown...")
					if agentEndOnError {
						if err := internal.StopAgentIfNeeded(); err != nil {
							failedStopAgent := true
							if isAdmin {
								_, err := commonProcess.Kill(common.GetExeFileName(true, common.LauncherConfigAdminAgent))
								if err == nil {
									fmt.Println("Successfully killed 'config-admin-agent'.")
									failedStopAgent = false
								}
							}
							if failedStopAgent {
								fmt.Println("Failed to stop 'config-admin-agent'. Kill it manually using the task manager")
								fmt.Println("Error message: " + err.Error())
								os.Exit(internal.ErrStartAgentRevert)
							}
						} else {
							fmt.Println("Successfully stopped 'config-admin-agent'.")
						}
					}
				} else {
					fmt.Println("Failed to run 'config-admin'")
				}
				os.Exit(errorCode)
			}
		}
	},
}

func InitSetUp() {
	if runtime.GOOS != "linux" {
		storeString = "user/" + storeString
	}
	cmd.InitSetUp(setUpCmd)
	commonCmd.GameVarCommand(setUpCmd.Flags(), &gameId)
	setUpCmd.Flags().StringVarP(
		&hostFilePath,
		"hostFilePath",
		"o",
		"",
		"Path to the host file. Only relevant when using 'ip' and/or 'CDN' option. If empty, it will use the system path",
	)
	setUpCmd.Flags().StringVarP(
		&certFilePath,
		"certFilePath",
		"t",
		"",
		"Path to the certificate file. It requires the 'localCert' option to be set. If non-empty the certificate will be saved only to the specified path.",
	)
	if runtime.GOOS != "linux" {
		setUpCmd.Flags().BytesBase64VarP(
			&AddUserCertData,
			"userCert",
			"u",
			nil,
			"Add the certificate to the user's trusted root store",
		)
	}
	if runtime.GOOS != "windows" {
		setUpCmd.Flags().StringVarP(
			&windowsUserProfilePath,
			"windowsUserProfilePath",
			"s",
			"",
			"Windows User Profile Path. Only relevant when using the 'metadata' or 'profiles' option.",
		)
	}
	setUpCmd.Flags().BoolVarP(
		&BackupMetadata,
		"metadata",
		"m",
		false,
		"Backup metadata. Not compatible with AoE:DE",
	)
	setUpCmd.Flags().BoolVarP(
		&BackupProfiles,
		"profiles",
		"p",
		false,
		"Backup profiles",
	)
	setUpCmd.Flags().BoolVarP(
		&agentStart,
		"agentStart",
		"g",
		false,
		"Start the 'config-admin-agent' if it is not running, we are not admin and is needed for admin action.",
	)
	setUpCmd.Flags().BoolVarP(
		&agentEndOnError,
		"agentEndOnError",
		"r",
		false,
		"Stop the 'config-admin-agent' if it is running and any admin action failed.",
	)
	err := setUpCmd.Flags().MarkHidden("agentStart")
	if err != nil {
		panic(err)
	}
	err = setUpCmd.Flags().MarkHidden("agentEndOnError")
	if err != nil {
		panic(err)
	}
	RootCmd.AddCommand(setUpCmd)
}
