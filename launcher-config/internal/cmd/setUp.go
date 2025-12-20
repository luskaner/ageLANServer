package cmd

import (
	"crypto/x509"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/cert"
	"github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
	"github.com/luskaner/ageLANServer/launcher-config/internal/admin"
	"github.com/luskaner/ageLANServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/ageLANServer/launcher-config/internal/userData"
	"github.com/spf13/cobra"
)

func removeUserCert() bool {
	commonLogger.Println("Removing previously added user certificate, authorize it if needed ...")
	if _, err := wrapper.RemoveUserCerts(); err == nil {
		commonLogger.Println("Successfully removed user certificate")
		return true
	}

	commonLogger.Println("Failed to remove user certificate")
	return false
}

func restoreMetadata() bool {
	commonLogger.Println("Restoring previously backed up metadata")
	if userData.Metadata(cmd.GameId).Restore() {
		commonLogger.Println("Successfully restored metadata")
		return true
	}

	commonLogger.Println("Failed to restore metadata")
	return false
}

func restoreProfiles() bool {
	commonLogger.Println("Restoring previously backed up profiles")
	if userData.RestoreProfiles(cmd.GameId, true) {
		commonLogger.Println("Successfully restored profiles")
		return true
	}

	commonLogger.Println("Failed to restore profiles")
	return false
}

func restoreGameCert() bool {
	commonLogger.Println("Restoring previously added game's certificate store...")
	if _, err := internal.NewCACert(cmd.GameId, gamePath).Restore(); err == nil {
		commonLogger.Println("Successfully restored game's certificate store.")
		return true
	}

	commonLogger.Println("Failed to restore game's certificate store.")
	return false
}

func undoSetUp() {
	if addedUserCert {
		removeUserCert()
	}
	if backedUpMetadata {
		restoreMetadata()
	}
	if backedUpProfiles {
		restoreProfiles()
	}
	if addedGameCert {
		restoreGameCert()
	}
	if hostFilePath != "" {
		_ = os.Remove(hostFilePath)
	}
	if certFilePath != "" {
		_ = os.Remove(certFilePath)
	}
	os.Exit(errorCode)
}

var addUserCertData []byte
var doBackupMetadata bool
var doBackupProfiles bool
var caStoreCert []byte
var agentStart bool
var agentEndOnError bool
var errorCode int
var storeString = "local"

// State
var addedUserCert bool
var backedUpMetadata bool
var backedUpProfiles bool
var addedGameCert bool

var setUpCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups configuration",
	Long:  "Adds any of the following:\n* One or more host mappings to the local DNS server\n* Certificate to the " + storeString + " machine's trusted root store\n* Backup user metadata\n* Backup user profiles",
	Run: func(_ *cobra.Command, _ []string) {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			_, ok := <-sigs
			if ok {
				errorCode = common.ErrSignal
				undoSetUp()
			}
		}()
		if logRoot != "" {
			internal.Initialize(logRoot)
		}
		if cmd.GameId == common.GameAoE1 {
			doBackupMetadata = false
			doRestoreCaStoreCert = false
		}
		if (doBackupMetadata || doBackupProfiles) && !common.SupportedGames.ContainsOne(cmd.GameId) {
			commonLogger.Println("Invalid game type")
			os.Exit(launcherCommon.ErrInvalidGame)
		}
		var addLocalCertData []byte = nil
		if certFilePath != "" {
			if len(cmd.AddLocalCertData) == 0 {
				commonLogger.Println("Certificate file path is set but no local certificate data is provided")
				os.Exit(internal.ErrMissingLocalCertData)
			}
		} else {
			addLocalCertData = cmd.AddLocalCertData
		}
		commonLogger.Printf("Setting up configuration for %s...\n", cmd.GameId)
		isAdmin := executor.IsAdmin()
		if addUserCertData != nil {
			commonLogger.Println("Adding user certificate, authorize it if needed...")
			crt := wrapper.BytesToCertificate(addUserCertData)
			if crt == nil {
				commonLogger.Println("Failed to parse certificate")
				errorCode = internal.ErrUserCertAddParse
				undoSetUp()
			}
			if err := wrapper.AddUserCerts([]*x509.Certificate{crt}); err == nil {
				commonLogger.Println("Successfully added user certificate")
				addedUserCert = true
			} else {
				commonLogger.Println("Failed to add user certificate")
				commonLogger.Println("Error message: " + err.Error())
				errorCode = internal.ErrUserCertAdd
				undoSetUp()
			}
		}
		if doBackupMetadata {
			commonLogger.Println("Backing up metadata")
			if userData.Metadata(cmd.GameId).Backup() {
				commonLogger.Println("Successfully backed up metadata")
				backedUpMetadata = true
			} else {
				commonLogger.Println("Failed to back up metadata")
				errorCode = internal.ErrMetadataBackup
				undoSetUp()
			}
		}
		if doBackupProfiles {
			commonLogger.Println("Backing up profiles")
			if userData.BackupProfiles(cmd.GameId) {
				commonLogger.Println("Successfully backed up profiles")
				backedUpProfiles = true
			} else {
				commonLogger.Println("Failed to back up profiles")
				errorCode = internal.ErrProfilesBackup
				undoSetUp()
			}
		}
		if caStoreCert != nil {
			commonLogger.Println("Adding certificate to game's store...")
			if gamePath == "" {
				commonLogger.Println("Game path is required to add certificate to game's store")
				errorCode = internal.ErrGamePathMissing
				undoSetUp()
			}
			crt := wrapper.BytesToCertificate(caStoreCert)
			if crt == nil {
				commonLogger.Println("Failed to parse certificate")
				errorCode = internal.ErrGameCertAddParse
				undoSetUp()
			}
			cert := internal.NewCACert(cmd.GameId, gamePath)
			if err := cert.Backup(); err == nil {
				commonLogger.Println("Successfully backed up game's store.")
				addedGameCert = true
			} else {
				commonLogger.Println("Failed to add certificate to game's store.")
				commonLogger.Println("Error message: " + err.Error())
				errorCode = internal.ErrGameCertBackup
				undoSetUp()
			}
			if err := cert.Append([]*x509.Certificate{crt}); err == nil {
				commonLogger.Println("Successfully added certificate to game's store.")
			} else {
				commonLogger.Println("Failed to add certificate to game's store.")
				commonLogger.Println("Error message: " + err.Error())
				errorCode = internal.ErrGameCertAdd
				undoSetUp()
			}
		}
		var ipToMap net.IP
		if hostFilePath == "" {
			if len(cmd.MapIP) > 0 {
				ipToMap = cmd.MapIP
			}
		} else if len(cmd.MapIP) > 0 {
			if ok, _ := hosts.AddHosts(cmd.GameId, hostFilePath, hosts.WindowsLineEnding, nil); ok {
				commonLogger.Println("Successfully added host mappings")
			} else {
				commonLogger.Println("Failed to add host mappings")
				errorCode = internal.ErrHostsAdd
				undoSetUp()
			}
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
				commonLogger.Println("Error saving certificate file:", err)
				errorCode = internal.ErrUserCertAdd
				undoSetUp()
			}
		}
		if addLocalCertData != nil || len(ipToMap) > 0 {
			agentStarted := admin.ConnectAgentIfNeeded() == nil
			if !agentStarted && agentStart && !isAdmin {
				result := admin.StartAgentIfNeeded()
				if !result.Success() {
					commonLogger.Println("Failed to start 'config-admin-agent'")
					if result.Err != nil {
						commonLogger.Println(result.Err)
					}
					if result.ExitCode != common.ErrSuccess {
						commonLogger.Println(result.ExitCode)
					}
					errorCode = internal.ErrStartAgent
					undoSetUp()
				} else {
					agentStarted = admin.ConnectAgentIfNeededWithRetries(true)
					if !agentStarted {
						commonLogger.Println("Failed to connect to 'config-admin-agent' after starting it. Kill it using the task manager.")
						errorCode = internal.ErrStartAgentVerify
						undoSetUp()
					}
				}
			}
			if agentStarted {
				commonLogger.Println("Communicating with 'config-admin-agent' to add local cert and/or host mappings...")
			} else {
				str := "Running 'config-admin' to add local cert and/or host mappings"
				if !isAdmin {
					str += ", authorize it if needed"
				}
				commonLogger.Println(str + "...")
			}
			err, exitCode := admin.RunSetUp(logRoot, ipToMap, addLocalCertData)
			if err == nil && exitCode == common.ErrSuccess {
				if agentStarted {
					commonLogger.Println("Successfully communicated with 'config-admin-agent'")
				} else {
					commonLogger.Println("Successfully ran 'config-admin'")
				}
			} else {
				if err != nil {
					commonLogger.Println("Received error:")
					commonLogger.Println(err)
				}
				if exitCode != common.ErrSuccess {
					commonLogger.Println("Received exit code:")
					commonLogger.Println(exitCode)
				}
				errorCode = internal.ErrAdminSetup
				if agentStarted {
					commonLogger.Println("Failed to communicate with 'config-admin-agent'. Communicating with it to shutdown...")
					if agentEndOnError {
						if err := admin.StopAgentIfNeeded(); err != nil {
							failedStopAgent := true
							if isAdmin {
								err := commonProcess.Kill(executables.Filename(true, executables.LauncherConfigAdminAgent))
								if err == nil {
									commonLogger.Println("Successfully killed 'config-admin-agent'.")
									failedStopAgent = false
								}
							}
							if failedStopAgent {
								commonLogger.Println("Failed to stop 'config-admin-agent'. Kill it manually using the task manager")
								commonLogger.Println("Error message: " + err.Error())
							}
						} else {
							commonLogger.Println("Successfully stopped 'config-admin-agent'.")
						}
					}
				} else {
					commonLogger.Println("Failed to run 'config-admin'")
				}
				undoSetUp()
			}
		}
	},
}

func InitSetUp() {
	if runtime.GOOS != "linux" {
		storeString = "user/" + storeString
	}
	cmd.InitSetUp(setUpCmd)
	addGamePathFlag(setUpCmd)
	commonCmd.LogRootCommand(setUpCmd.Flags(), &logRoot)
	setUpCmd.Flags().StringVarP(
		&hostFilePath,
		"hostFilePath",
		"o",
		"",
		"Path to the host file. Only relevant when using 'ip' option. If empty, it will use the system path",
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
			&addUserCertData,
			"userCert",
			"u",
			nil,
			"Add the certificate to the user's trusted root store",
		)
	}
	setUpCmd.Flags().BoolVarP(
		&doBackupMetadata,
		"metadata",
		"m",
		false,
		"Backup metadata. Not compatible with AoE:DE",
	)
	setUpCmd.Flags().BoolVarP(
		&doBackupProfiles,
		"profiles",
		"p",
		false,
		"Backup profiles",
	)
	setUpCmd.Flags().BytesBase64VarP(
		&caStoreCert,
		"caStoreCert",
		"s",
		nil,
		"Add the certificate to the game's trusted root store. For all except AoE I: DE.",
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
