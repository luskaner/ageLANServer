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
	ErrConfigIpMapAdd
	ErrConfigCertAdd
	ErrConfigCert
	ErrReadCert
	ErrTrustCert
	ErrMetadataSetup
	ErrAgentStart
	ErrInvalidServerArgs
	ErrInvalidClientArgs
	ErrInvalidSetupCommand
	ErrInvalidRevertCommand
	ErrSetupCommand
	ErrConfigCDNMap
	ErrSteamRoot
	ErrAnnouncementMulticastGroup
	ErrCertMismatch
	ErrInvalidIsolationWindowsUserProfilePath
	ErrServerStartDeclined
	ErrAnnouncementPort
	ErrAnnouncement
)
