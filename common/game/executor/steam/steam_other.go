//go:build !darwin

package steam

import (
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game/executor/base"
	"github.com/luskaner/ageLANServer/common/game/steam"
)

func (exec Exec) Do(_ []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	return base.StartUri(exec.OpenUri(), optionsFn)
}

func NewExec(gameId string) (exec *Exec, ok bool) {
	var game *steam.Game
	if game, ok = steam.NewGame(gameId); ok {
		exec, ok = NewExecFromGame(game)
	}
	return
}
