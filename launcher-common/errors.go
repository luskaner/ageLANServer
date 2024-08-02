package launcher_common

import "github.com/luskaner/aoe2DELanServer/common"

const (
	ErrNotAdmin = iota + common.ErrLast
	ErrIpMapAddTooMany
	// ErrLast Only used as a marker to where to start
	ErrLast
)
