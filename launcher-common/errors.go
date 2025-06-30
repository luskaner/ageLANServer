package launcher_common

import "github.com/luskaner/ageLANServer/common"

const (
	ErrNotAdmin = iota + common.ErrLast
	ErrInvalidGameTitle
	// ErrLast Only used as a marker to where to start
	ErrLast
)
