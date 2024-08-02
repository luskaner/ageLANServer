package internal

import (
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
)

const (
	ErrLocalCertRemove = iota + launcherCommon.ErrLast
	ErrIpMapRemove
	ErrIpMapRemoveRevert
	ErrLocalCertAdd
	ErrLocalCertAddParse
	ErrIpMapAdd
	ErrIpMapAddRevert
)
