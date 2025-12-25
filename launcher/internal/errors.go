package internal

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

const (
	ErrInvalidCanTrustCertificate = iota + launcherCommon.ErrLast
	ErrInvalidCanBroadcastBattleServer
	ErrInvalidServerStart
	ErrInvalidServerStop
	ErrInvalidServerHost
	ErrGameAlreadyRunning
	ErrGameLauncherNotFound
	ErrGameLauncherStart
	ErrServerExecutable
	ErrServerConnectSecure
	ErrServerUnreachable
	ErrServerCertMissingExpired
	ErrServerCertDirectory
	ErrServerCertCreate
	ErrServerStart
	ErrConfigIpMap
	ErrGameUnsupportedLauncherCombo
	ErrConfigIpMapAdd
	ErrConfigCertAdd
	ErrConfigCACertAdd
	ErrConfigCert
	ErrReadCert
	ErrTrustCert
	ErrMetadataProfilesSetup
	ErrAgentStart
	ErrInvalidClientPath
	ErrInvalidServerArgs
	ErrInvalidServerPath
	ErrInvalidClientArgs
	ErrInvalidSetupCommand
	ErrInvalidRevertCommand
	ErrSetupCommand
	ErrSteamRoot
	ErrAnnouncementMulticastGroup
	ErrCertMismatch
	ErrInvalidServerBattleServerManagerRun
	ErrInvalidServerBattleServerManagerArgs
	ErrBattleServerManagerRun
	ErrInvalidIsolateMetadata
	ErrInvalidIsolateProfiles
	ErrRequiredIsolation
	ErrGameConfigParse
)
