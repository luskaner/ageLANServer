package internal

import (
	"github.com/luskaner/ageLANServer/common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrResolveHost
	ErrNoAddrs
	ErrCreateLogsDir
	ErrCreateLogFile
	ErrStartServer
	ErrQueryServer
)
