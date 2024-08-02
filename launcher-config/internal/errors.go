package internal

import (
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
)

const (
	ErrUserCertRemove = iota + launcherCommon.ErrLast
	ErrUserCertAdd
	ErrUserCertAddParse
	ErrMetadataRestore
	ErrMetadataRestoreRevert
	ErrProfilesRestore
	ErrProfilesRestoreRevert
	ErrAdminRevert
	ErrAdminRevertRevert
	ErrMetadataBackup
	ErrMetadataBackupRevert
	ErrProfilesBackup
	ErrProfilesBackupRevert
	ErrStopAgent
	ErrStopAgentVerify
	ErrStartAgent
	ErrStartAgentRevert
	ErrStartAgentVerify
	ErrAdminSetup
	ErrAdminSetupRevert
)
