package cmd

import (
	"crypto/x509"
	"encoding/base64"
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
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/hosts"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
	"github.com/luskaner/ageLANServer/launcher-config/internal/admin"
	"github.com/luskaner/ageLANServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/ageLANServer/launcher-config/internal/userData"
	"github.com/spf13/pflag"
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
	if userData.Metadata(launcherCommonCmd.GameId).Restore() {
		commonLogger.Println("Successfully restored metadata")
		return true
	}

	commonLogger.Println("Failed to restore metadata")
	return false
}

func restoreProfiles() bool {
	commonLogger.Println("Restoring previously backed up profiles")
	if userData.RestoreProfiles(launcherCommonCmd.GameId, true) {
		commonLogger.Println("Successfully restored profiles")
		return true
	}

	commonLogger.Println("Failed to restore profiles")
	return false
}

func restoreGameCert() bool {
	commonLogger.Println("Restoring previously added game's certificate store...")
	if _, err := internal.NewCACert(launcherCommonCmd.GameId, gamePath).Restore(); err == nil {
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
var addUserCertDataB64 string
var doBackupMetadata bool
var doBackupProfiles bool
var caStoreCert []byte
var caStoreCertB64 string
var agentStart bool
var agentEndOnError bool
var errorCode int

// State
var addedUserCert bool
var backedUpMetadata bool
var backedUpProfiles bool
var addedGameCert bool

func runSetUp(args []string) error {
	fs := pflag.NewFlagSet("setup", pflag.ContinueOnError)
	launcherCommonCmd.InitSetUp(fs)
	addGamePathFlag(fs)
	commonCmd.LogRootCommand(fs, &logRoot)
	fs.StringVarP(&hostFilePath, "hostFilePath", "o", "", "Path to the host file. Only relevant when using 'ip' option. If empty, it will use the system path")
	fs.StringVarP(&certFilePath, "certFilePath", "t", "", "Path to the certificate file. It requires the 'localCert' option to be set. If non-empty the certificate will be saved only to the specified path.")
	if runtime.GOOS != "linux" {
		fs.StringVarP(&addUserCertDataB64, "userCert", "u", "", "Add the certificate to the user's trusted root store")
	}
	fs.BoolVarP(&doBackupMetadata, "metadata", "m", false, "Backup metadata. Not compatible with AoE:DE")
	fs.BoolVarP(&doBackupProfiles, "profiles", "p", false, "Backup profiles")
	fs.StringVarP(&caStoreCertB64, "caStoreCert", "s", "", "Add the certificate to the game's trusted root store. For all except AoE I: DE and AoE IV: AE.")
	fs.BoolVarP(&agentStart, "agentStart", "g", false, "Start the 'config-admin-agent' if it is not running, we are not admin and is needed for admin action.")
	fs.BoolVarP(&agentEndOnError, "agentEndOnError", "r", false, "Stop the 'config-admin-agent' if it is running and any admin action failed.")
	_ = fs.MarkHidden("agentStart")
	_ = fs.MarkHidden("agentEndOnError")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// decode base64 flags
	if err := launcherCommonCmd.DecodeSetUpFlags(); err != nil {
		return err
	}
	if addUserCertDataB64 != "" {
		if b, err := base64.StdEncoding.DecodeString(addUserCertDataB64); err == nil {
			addUserCertData = b
		} else {
			return err
		}
	}
	if caStoreCertB64 != "" {
		if b, err := base64.StdEncoding.DecodeString(caStoreCertB64); err == nil {
			caStoreCert = b
		} else {
			return err
		}
	}

	// signal handler
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
	if launcherCommonCmd.GameId == common.GameAoE1 {
		doBackupMetadata = false
		doRestoreCaStoreCert = false
	} else if launcherCommonCmd.GameId == common.GameAoE4 {
		doRestoreCaStoreCert = false
	}
	if (doBackupMetadata || doBackupProfiles) && !common.SupportedGames.ContainsOne(launcherCommonCmd.GameId) {
		commonLogger.Println("Invalid game type")
		os.Exit(launcherCommon.ErrInvalidGame)
	}
	var addLocalCertData []byte = nil
	if certFilePath != "" {
		if len(launcherCommonCmd.AddLocalCertData) == 0 {
			commonLogger.Println("Certificate file path is set but no local certificate data is provided")
			os.Exit(internal.ErrMissingLocalCertData)
		}
	} else {
		addLocalCertData = launcherCommonCmd.AddLocalCertData
	}
	commonLogger.Printf("Setting up configuration for %s...\n", launcherCommonCmd.GameId)
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
		if userData.Metadata(launcherCommonCmd.GameId).Backup() {
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
		if userData.BackupProfiles(launcherCommonCmd.GameId) {
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
		gameCert := internal.NewCACert(launcherCommonCmd.GameId, gamePath)
		if err := gameCert.Backup(); err == nil {
			commonLogger.Println("Successfully backed up game's store.")
			addedGameCert = true
		} else {
			commonLogger.Println("Failed to add certificate to game's store.")
			commonLogger.Println("Error message: " + err.Error())
			errorCode = internal.ErrGameCertBackup
			undoSetUp()
		}
		if err := gameCert.Append([]*x509.Certificate{crt}); err == nil {
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
		if len(launcherCommonCmd.MapIP) > 0 {
			ipToMap = launcherCommonCmd.MapIP
		}
	} else if len(launcherCommonCmd.MapIP) > 0 {
		if ok, _ := hosts.AddHosts(launcherCommonCmd.GameId, hostFilePath, hosts.WindowsLineEnding, nil); ok {
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
			err = cert.WriteAsPem(launcherCommonCmd.AddLocalCertData, certFile)
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
				} else {
					commonLogger.Println("Failed to run 'config-admin'")
				}
				undoSetUp()
			}
		}
	}
	return nil
}
