package internal

import (
	"github.com/luskaner/ageLANServer/common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrResolveHost
	ErrCreateLogsDir
	ErrCreateLogFile
	ErrStartServer
	ErrMulticastGroup
	ErrGames
	ErrInvalidId
	ErrAnnouncePort
)
