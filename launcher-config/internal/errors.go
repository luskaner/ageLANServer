package internal

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

const (
	ErrUserCertRemove = iota + launcherCommon.ErrLast
	ErrUserCertAdd
	ErrUserCertAddParse
	ErrMetadataRestore
	ErrProfilesRestore
	ErrAdminRevert
	ErrMetadataBackup
	ErrProfilesBackup
	ErrStartAgent
	ErrStartAgentVerify
	ErrAdminSetup
	ErrRevertStopAgent
	ErrHostsAdd
	ErrMissingLocalCertData
	ErrGameCertAddParse
	ErrGameCertAdd
	ErrGameCertRestore
	ErrGameCertBackup
	ErrGamePathMissing
)
