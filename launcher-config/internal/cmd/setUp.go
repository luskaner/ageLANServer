package cmd

import (
	"crypto/x509"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/common/hosts"
	"github.com/luskaner/ageLANServer/common/logger"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	launcherCommonCmd "github.com/luskaner/ageLANServer/launcher-common/cmd/config"
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
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
	if userData.Metadata(path).Restore() {
		commonLogger.Println("Successfully restored metadata")
		return true
	}

	commonLogger.Println("Failed to restore metadata")
	return false
}

func restoreProfiles() bool {
	commonLogger.Println("Restoring previously backed up profiles")
	if userData.RestoreProfiles(path, true) {
		commonLogger.Println("Successfully restored profiles")
		return true
	}

	commonLogger.Println("Failed to restore profiles")
	return false
}

func restoreGameCert() bool {
	commonLogger.Println("Restoring previously added game's certificate store...")
	if _, err := internal.NewCACert(setupValues.GameId, setupValues.GamePath).Restore(); err == nil {
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
	if setupValues.HostFilePath != "" {
		_ = os.Remove(setupValues.HostFilePath)
	}
	if setupValues.CertFilePath != "" {
		_ = os.Remove(setupValues.CertFilePath)
	}
}

var setupValues *launcherCommonCmd.SetupValues

// State
var addedUserCert bool
var backedUpMetadata bool
var backedUpProfiles bool
var addedGameCert bool

func runSetUp(args []string) (err error, exitCode int) {
	var flags *pflag.FlagSet
	setupValues, flags = launcherCommonCmd.SetUpFlagSet()
	if err = flags.Parse(args); err != nil {
		exitCode = common.ErrSyntax
		return
	}

	if setupValues.GameId == "" {
		return errors.New("required flag 'game' not set"), common.ErrSyntax
	}

	// signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_, ok := <-sigs
		if ok {
			exitCode = common.ErrSignal
			undoSetUp()
		}
	}()
	if setupValues.LogRoot != "" {
		internal.Initialize(setupValues.LogRoot)
	}
	if setupValues.GameId == game.AoE1 {
		setupValues.Metadata = false
		setupValues.AddCACertData = nil
	} else if setupValues.GameId == game.AoE4 {
		setupValues.AddCACertData = nil
	}
	if setupValues.Metadata || setupValues.Profiles {
		var fileInfo os.FileInfo
		if !game.SupportedGames.ContainsOne(setupValues.GameId) {
			commonLogger.Println("Invalid game type")
			exitCode = launcherCommon.ErrInvalidGame
			return
		} else if fileInfo, err = os.Stat(setupValues.DataPath); err != nil || !fileInfo.IsDir() {
			commonLogger.Println("Invalid data path")
			exitCode = internal.ErrInvalidDataPath
			return
		}
		path = commonUserData.NewPath(setupValues.DataPath, setupValues.GameId)
	}

	var addLocalCertData []byte = nil
	if setupValues.CertFilePath != "" {
		if len(setupValues.AddLocalCertData) == 0 {
			commonLogger.Println("Certificate file path is set but no local certificate data is provided")
			exitCode = internal.ErrMissingLocalCertData
			return
		}
	} else {
		addLocalCertData = setupValues.AddLocalCertData
	}
	commonLogger.Printf("Setting up configuration for %s...\n", setupValues.GameId)
	isAdmin := executor.IsAdmin()
	if setupValues.AddUserCertData != nil {
		commonLogger.Println("Adding user certificate, authorize it if needed...")
		crt := common.BytesToCertificate(setupValues.AddUserCertData)
		if crt == nil {
			commonLogger.Println("Failed to parse certificate")
			exitCode = internal.ErrUserCertAddParse
			undoSetUp()
			return
		}
		if err = wrapper.AddUserCerts([]*x509.Certificate{crt}); err == nil {
			commonLogger.Println("Successfully added user certificate")
			addedUserCert = true
		} else {
			commonLogger.Println("Failed to add user certificate")
			commonLogger.Println("Error message: " + err.Error())
			exitCode = internal.ErrUserCertAdd
			undoSetUp()
			return
		}
	}
	if setupValues.Metadata {
		commonLogger.Println("Backing up metadata")
		if userData.Metadata(path).Backup() {
			commonLogger.Println("Successfully backed up metadata")
			backedUpMetadata = true
		} else {
			commonLogger.Println("Failed to back up metadata")
			exitCode = internal.ErrMetadataBackup
			undoSetUp()
			return
		}
	}
	if setupValues.Profiles {
		commonLogger.Println("Backing up profiles")
		if userData.BackupProfiles(path) {
			commonLogger.Println("Successfully backed up profiles")
			backedUpProfiles = true
		} else {
			commonLogger.Println("Failed to back up profiles")
			exitCode = internal.ErrProfilesBackup
			undoSetUp()
			return
		}
	}
	if setupValues.AddCACertData != nil {
		commonLogger.Println("Adding certificate to game's store...")
		if setupValues.GamePath == "" {
			commonLogger.Println("Game path is required to add certificate to game's store")
			exitCode = internal.ErrGamePathMissing
			undoSetUp()
			return
		}
		crt := common.BytesToCertificate(setupValues.AddCACertData)
		if crt == nil {
			commonLogger.Println("Failed to parse certificate")
			exitCode = internal.ErrGameCertAddParse
			undoSetUp()
			return
		}
		gameCert := internal.NewCACert(setupValues.GameId, setupValues.GamePath)
		if err = gameCert.Backup(); err == nil {
			commonLogger.Println("Successfully backed up game's store.")
			addedGameCert = true
		} else {
			commonLogger.Println("Failed to add certificate to game's store.")
			commonLogger.Println("Error message: " + err.Error())
			exitCode = internal.ErrGameCertBackup
			undoSetUp()
			return
		}
		if err = gameCert.Append([]*x509.Certificate{crt}); err == nil {
			commonLogger.Println("Successfully added certificate to game's store.")
		} else {
			commonLogger.Println("Failed to add certificate to game's store.")
			commonLogger.Println("Error message: " + err.Error())
			exitCode = internal.ErrGameCertAdd
			undoSetUp()
			return
		}
	}
	var ipToMap net.IP
	if setupValues.HostFilePath == "" {
		if len(setupValues.MapIp) > 0 {
			ipToMap = setupValues.MapIp
		}
	} else if len(setupValues.MapIp) > 0 {
		if ok, _ := hosts.AddHosts(setupValues.MapIp, setupValues.GameId, setupValues.HostFilePath, hosts.WindowsLineEnding, setupValues.MacOsExclusiveMappings, nil); ok {
			commonLogger.Println("Successfully added host mappings")
		} else {
			commonLogger.Println("Failed to add host mappings")
			exitCode = internal.ErrHostsAdd
			undoSetUp()
			return
		}
	}
	if setupValues.CertFilePath != "" {
		var certFile *os.File
		certFile, err = os.Create(setupValues.CertFilePath)
		if err == nil {
			err = common.WriteAsPem(setupValues.AddLocalCertData, certFile)
			_ = certFile.Close()
		}
		if err != nil {
			commonLogger.Println("Error saving certificate file:", err)
			exitCode = internal.ErrUserCertAdd
			undoSetUp()
			return
		}
	}
	if addLocalCertData != nil || len(ipToMap) > 0 {
		agentStarted := admin.ConnectAgentIfNeeded() == nil
		if agentStarted {
			commonLogger.Println("Communicating with 'config-admin-agent' to add local cert and/or host mappings...")
		} else {
			str := "Running 'config-admin' to add local cert and/or host mappings"
			if !isAdmin {
				str += ", authorize it if needed"
			}
			commonLogger.Println(str + "...")
		}
		err, exitCode = admin.RunSetUp(setupValues.GameId, setupValues.LogRoot, ipToMap, setupValues.MacOsExclusiveMappings, addLocalCertData)
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
			exitCode = internal.ErrAdminSetup
			if agentStarted {
				commonLogger.Println("Failed to communicate with 'config-admin-agent'. Communicating with it to shutdown...")
				_ = admin.StopAgentIfNeeded()
				undoSetUp()
			}
		}
	}
	return
}
