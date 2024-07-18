package internal

import (
	"common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrCertCreate
	ErrCertCreateExisting
)
