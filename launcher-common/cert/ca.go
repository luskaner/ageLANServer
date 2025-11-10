package cert

import (
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
)

type CA struct {
	gamePath string
}

func NewCA(gameId string, gamePath string) CA {
	if gameId == common.GameAoE2 {
		gamePath = filepath.Join(gamePath, "certificates")
	}
	return CA{gamePath: gamePath}
}

func (c CA) name() string {
	return "cacert.pem"
}

func (c CA) OriginalPath() string {
	return filepath.Join(c.gamePath, c.name())
}

func (c CA) TmpPath() string {
	return filepath.Join(c.gamePath, c.name()+".lan")
}

func (c CA) BackupPath() string {
	return filepath.Join(c.gamePath, c.name()+".bak")
}
