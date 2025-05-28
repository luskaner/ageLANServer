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
	ErrAdminRevert
	ErrAdminRevertRevert
	ErrMetadataBackup
	ErrMetadataBackupRevert
	ErrStartAgent
	ErrStartAgentRevert
	ErrStartAgentVerify
	ErrAdminSetup
	ErrAdminSetupRevert
	ErrRevertStopAgent
	ErrHostsAdd
	ErrMissingLocalCertData
)
