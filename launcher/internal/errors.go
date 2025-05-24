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
	ErrListenServerAnnouncements
	ErrServerExecutable
	ErrServerConnectSecure
	ErrServerUnreachable
	ErrServerCertMissingExpired
	ErrServerCertDirectory
	ErrServerCertCreate
	ErrServerStart
	ErrConfigIpMap
	ErrConfigIpMapFind
	ErrConfigIpMapAdd
	ErrConfigCertAdd
	ErrConfigCert
	ErrReadCert
	ErrTrustCert
	ErrMetadataProfilesSetup
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
	ErrInvalidIsolationUserProfilePath
)
