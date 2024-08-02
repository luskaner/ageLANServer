package internal

import (
	"github.com/luskaner/aoe2DELanServer/common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrResolveHost
	ErrCreateLogsDir
	ErrCreateLogFile
	ErrStartServer
)
