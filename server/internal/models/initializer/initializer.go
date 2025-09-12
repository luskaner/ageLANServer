package initializer

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/age1"
	"github.com/luskaner/ageLANServer/server/internal/models/age2"
	"github.com/luskaner/ageLANServer/server/internal/models/age3"
	"github.com/luskaner/ageLANServer/server/internal/models/athens"
)

var Games = map[string]models.Game{}

func InitializeGames(gameIds mapset.Set[string]) {
	for gameId := range gameIds.Iter() {
		var game models.Game
		switch gameId {
		case common.GameAoE1:
			game = age1.CreateGame()
		case common.GameAoE2:
			game = age2.CreateGame()
		case common.GameAoE3:
			game = age3.CreateGame()
		case common.GameAoM:
			game = athens.CreateGame()
		}
		Games[gameId] = game
	}
}
