package internal

import (
	"github.com/luskaner/ageLANServer/common"
)

const (
	ErrCertDirectory = iota + common.ErrLast
	ErrResolveHost
	ErrCreateLogFile
	ErrStartServer
	ErrMulticastGroup
	ErrGames
	ErrGame
	ErrAnnounce
	ErrInvalidId
)
