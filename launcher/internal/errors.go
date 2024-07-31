package internal

import (
	launcherCommon "launcherCommon"
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
	ErrAnnouncementPort
	ErrListenServerAnnouncements
	ErrServerExecutable
	ErrServerConnectSecure
	ErrServerUnreachable
	ErrServerCertMissing
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
)
