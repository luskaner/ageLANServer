package internal

import (
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
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
	ErrStartAgent
	ErrStartAgentRevert
	ErrStartAgentVerify
	ErrAdminSetup
	ErrAdminSetupRevert
)
