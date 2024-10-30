package initializer

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/models/age2"
	"github.com/luskaner/aoe2DELanServer/server/internal/models/age3"
)

var Games = map[string]models.Game{}

func InitializeGames(gameIds mapset.Set[string]) {
	for gameId := range gameIds.Iter() {
		var game models.Game
		switch gameId {
		case common.GameAoE2:
			game = age2.CreateGame()
		case common.GameAoE3:
			game = age3.CreateGame()
		}
		Games[gameId] = game
	}
}
