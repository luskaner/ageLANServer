package internal

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

const (
	ErrInvalidServerStart = iota + launcherCommon.ErrLast
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
	ErrInvalidRevertCommand
	ErrSetupCommand
	ErrConfigCDNMap
	ErrCertMismatch
	ErrServerStartDeclined
)
