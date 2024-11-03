package internal

import (
	"github.com/luskaner/ageLANServer/common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrCertCreate
	ErrCertCreateExisting
)
