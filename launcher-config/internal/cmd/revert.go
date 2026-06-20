package cmd

import (
	"crypto/x509"
	"errors"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executables"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/logger"
	commonProcess "github.com/luskaner/ageLANServer/common/process"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
	"github.com/luskaner/ageLANServer/launcher-config/internal"
	"github.com/luskaner/ageLANServer/launcher-config/internal/admin"
	"github.com/luskaner/ageLANServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/ageLANServer/launcher-config/internal/userData"
	"github.com/spf13/pflag"
)

func addUserCerts(removedUserCerts []*x509.Certificate) bool {
	commonLogger.Println("Adding previously removed user certificate")
	if err := wrapper.AddUserCerts(removedUserCerts); err == nil {
		commonLogger.Println("Successfully added user certificate")
		return true
	}

	commonLogger.Println("Failed to add user certificate")
	return false
}

func backupMetadata() bool {
	commonLogger.Println("Backing up previously restored metadata")
	if userData.Metadata(path).Backup() {
		commonLogger.Println("Successfully backed up metadata")
		return true
	}

	commonLogger.Println("Failed to back up metadata")
	return false
}

func backupProfiles() bool {
	commonLogger.Println("Backing up previously restored profiles")
	if userData.BackupProfiles(path) {
		commonLogger.Println("Successfully backed up profiles")
		return true
	}

	commonLogger.Println("Failed to back up profiles")
	return false
}

func addCaCerts(removedCaCerts []*x509.Certificate) bool {
	commonLogger.Println("Restoring previously added game's certificate store...")
	if err := internal.NewCACert(revertValues.GameId, revertValues.GamePath).Append(removedCaCerts); err == nil {
		commonLogger.Println("Successfully restored game's certificate store.")
		return true
	}

	commonLogger.Println("Failed to restore game's certificate store.")
	return false
}

func undoRevert() {
	if !revertValues.RemoveAll {
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

var revertValues *launcherCommonCmd.RevertValues

// State
var removedUserCerts []*x509.Certificate
var removedCaCerts []*x509.Certificate
var restoredMetadata bool
var restoredProfiles bool

func runRevert(args []string) error {
	var flags *pflag.FlagSet
	revertValues, flags = launcherCommonCmd.RevertFlagSet()
	if err := flags.Parse(args); err != nil {
		return err
	}
	if revertValues.GameId == "" {
		return errors.New("required flag 'game' not set")
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			undoRevert()
			os.Exit(common.ErrSignal)
		}
	}()
	if revertValues.LogRoot != "" {
		internal.Initialize(revertValues.LogRoot)
	}
	isAdmin := executor.IsAdmin()
	reverseFailed := true
	if revertValues.RemoveAll {
		revertValues.IPs = true
		revertValues.Certs = true
		if runtime.GOOS != "linux" {
			revertValues.RemoveUserCert = true
		}
		revertValues.Metadata = true
		revertValues.Profiles = true
		revertValues.RestoreCAStoreCert = true
		reverseFailed = false
	}
	if revertValues.GameId == game.AoE1 {
		revertValues.Metadata = false
		revertValues.RestoreCAStoreCert = false
	} else if revertValues.GameId == game.AoE4 {
		revertValues.RestoreCAStoreCert = false
	}
	if revertValues.Metadata || revertValues.Profiles {
		if !game.SupportedGames.ContainsOne(revertValues.GameId) {
			commonLogger.Println("Invalid game type")
			errorCode = launcherCommon.ErrInvalidGame
			undoRevert()
		} else if fileInfo, err := os.Stat(revertValues.DataPath); err != nil || !fileInfo.IsDir() {
			commonLogger.Println("Invalid data path")
			errorCode = internal.ErrInvalidDataPath
			undoRevert()
		} else {
			path = commonUserData.NewPath(revertValues.DataPath, revertValues.GameId)
		}
	}
	commonLogger.Printf("Reverting configuration for %s...\n", revertValues.GameId)
	if revertValues.RemoveUserCert {
		commonLogger.Println("Removing user certificates, authorize it if needed...")
		if removedUserCerts, _ = wrapper.RemoveUserCerts(); removedUserCerts != nil {
			commonLogger.Println("Successfully removed user certificates")
		} else {
			commonLogger.Println("Failed to remove user certificates")
			errorCode = internal.ErrUserCertRemove
			undoRevert()
		}
	}
	if revertValues.Metadata {
		commonLogger.Println("Restoring metadata")
		if userData.Metadata(path).Restore() {
			commonLogger.Println("Successfully restored metadata")
			restoredMetadata = true
		} else {
			commonLogger.Println("Failed to restore metadata")
			errorCode = internal.ErrMetadataRestore
			undoRevert()
		}
	}
	if revertValues.Profiles {
		commonLogger.Println("Restoring profiles")
		if userData.RestoreProfiles(path, reverseFailed) {
			commonLogger.Println("Successfully restored profiles")
			restoredProfiles = true
		} else {
			commonLogger.Println("Failed to restore profiles")
			errorCode = internal.ErrProfilesRestore
			undoRevert()
		}
	}
	if revertValues.RestoreCAStoreCert {
		commonLogger.Println("Restoring original certificate game's store...")
		if revertValues.GamePath == "" {
			commonLogger.Println("Game path is required to restore the original game's store")
			errorCode = internal.ErrGamePathMissing
			undoRevert()
		}
		cert := internal.NewCACert(revertValues.GameId, revertValues.GamePath)
		var err error
		if err, removedCaCerts = cert.Restore(); err == nil {
			commonLogger.Println("Successfully restored original game's store.")
		} else {
			commonLogger.Println("Failed to restore original game's store.")
			commonLogger.Println("Received error:")
			commonLogger.Println(err)
			errorCode = internal.ErrGameCertRestore
			undoRevert()
		}
	}
	var agentConnected *bool
	if launcherCommon.RevertRequiresAdminElevationValues(revertValues) {
		agentConnected = new(admin.ConnectAgentIfNeeded() == nil)
		if *agentConnected {
			commonLogger.Println("Communicating with 'config-admin-agent' to remove local cert and/or host mappings...")
		} else {
			str := "Running 'config-admin' to remove local cert and/or host mappings"
			if !isAdmin {
				str += ", authorize it if needed"
			}
			commonLogger.Println(str + "...")
		}
		var err error
		err, errorCode = admin.RunRevert(revertValues.LogRoot, revertValues.IPs, revertValues.Certs, !revertValues.RemoveAll)
		if err == nil && errorCode == common.ErrSuccess {
			if *agentConnected {
				commonLogger.Println("Successfully communicated with 'config-admin-agent'")
			} else {
				commonLogger.Println("Successfully ran 'config-admin'")
			}
		} else {
			if err != nil {
				commonLogger.Println("Received error:")
				commonLogger.Println(err)
			}
			if errorCode != common.ErrSuccess {
				commonLogger.Println("Received exit code:")
				commonLogger.Println(errorCode)
			}
			errorCode = internal.ErrAdminRevert
			undoRevert()
			if *agentConnected {
				commonLogger.Println("Failed to communicate with 'config-admin-agent'")
			} else {
				commonLogger.Println("Failed to run 'config-admin'")
			}
		}
	}
	// Ignore previous error if we don't failfast
	if revertValues.RemoveAll {
		errorCode = common.ErrSuccess
	}
	if errorCode == common.ErrSuccess && revertValues.HostFilePath != "" {
		_ = os.Remove(revertValues.HostFilePath)
	}
	if errorCode == common.ErrSuccess && revertValues.CertFilePath != "" {
		_ = os.Remove(revertValues.CertFilePath)
	}
	if agentConnected != nil {
		failedStopAgent := true
		if *agentConnected {
			commonLogger.Println("Trying to stop 'config-admin-agent'.")
			err := admin.StopAgentIfNeeded()
			if err == nil {
				if admin.ConnectAgentIfNeededWithRetries(false) {
					commonLogger.Println("Stopped 'config-admin-agent'")
					failedStopAgent = false
				} else {
					commonLogger.Println("Failed to stop 'config-admin-agent'")
				}
			} else {
				commonLogger.Println("Failed to trying stopping 'config-admin-agent'")
				commonLogger.Println(err)
			}
		}
		if failedStopAgent {
			exeFileName := executables.NativeFileName(true, executables.LauncherConfigAdminAgent)
			if pid, proc, err := commonProcess.Process(exeFileName); err == nil && proc != nil {
				if isAdmin {
					if err := commonProcess.KillPidProc(pid, proc); err == nil {
						commonLogger.Println("Successfully killed 'config-admin-agent'.")
						failedStopAgent = false
					} else {
						commonLogger.Println("Failed to kill 'config-admin-agent'")
					}
				} else {
					commonLogger.Println("Re-run as administrator to kill 'config-admin-agent'")
				}
			} else if err == nil && proc == nil {
				failedStopAgent = false
			}
		}
		if failedStopAgent && errorCode == common.ErrSuccess {
			errorCode = internal.ErrRevertStopAgent
		}
	}
	os.Exit(errorCode)
	return nil
}
