package cmdUtils

import (
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

var GameIds []string

func ParsedGameIds(gameIds *[]string) (games mapset.Set[string], err error) {
	if gameIds == nil {
		gameIds = &GameIds
	}
	if len(*gameIds) == 0 {
		games = common.SupportedGames
	} else if !common.SupportedGames.IsSuperset(mapset.NewThreadUnsafeSet[string](*gameIds...)) {
		err = fmt.Errorf("game(s) not supported")
		return
	} else {
		games = mapset.NewSet[string](*gameIds...)
	}
	return
}
