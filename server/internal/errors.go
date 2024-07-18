package internal

import (
	"common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrResolveHost
	ErrCreateLogsDir
	ErrCreateLogFile
	ErrStartServer
)
