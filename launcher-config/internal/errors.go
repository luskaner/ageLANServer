package internal

import (
	launcherCommon "launcher-common"
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
