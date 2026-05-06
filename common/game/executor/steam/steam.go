package steam

import (
	"github.com/luskaner/ageLANServer/common/game/steam"
)

type Exec struct {
	*steam.Game
}

func (exec Exec) GameProcesses() (steamProcess bool, xboxProcess bool) {
	steamProcess = true
	return
}

func (exec Exec) String() string {
	return "Steam"
}

func NewExecFromGame(game *steam.Game) (exec *Exec, ok bool) {
	if game != nil {
		exec = &Exec{game}
		ok = true
	}
	return
}
