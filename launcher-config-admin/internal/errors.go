package internal

import (
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcherCommon"
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
