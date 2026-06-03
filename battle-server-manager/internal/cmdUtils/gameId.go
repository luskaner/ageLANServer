package cmdUtils

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/game"
)

var GameIds []string

func ParsedGameIds(gameIds *[]string) (games mapset.Set[string], err error) {
	if gameIds == nil {
		gameIds = &GameIds
	}
	if len(*gameIds) == 0 {
		games = game.SupportedGames
	} else if !game.SupportedGames.IsSuperset(mapset.NewThreadUnsafeSet[string](*gameIds...)) {
		err = fmt.Errorf("game(s) not supported")
		return
	}

	games = mapset.NewSet[string](*gameIds...)
	return
}
