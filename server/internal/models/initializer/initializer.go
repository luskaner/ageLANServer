package initializer

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/age1"
	"github.com/luskaner/ageLANServer/server/internal/models/age2"
	"github.com/luskaner/ageLANServer/server/internal/models/age3"
)

var GameTitles = map[common.GameTitle]models.Game{}

func InitializeGames(gameTitles mapset.Set[common.GameTitle]) {
	for gameTitle := range gameTitles.Iter() {
		var game models.Game
		switch gameTitle {
		case common.AoE1:
			game = age1.CreateGame()
		case common.AoE2:
			game = age2.CreateGame()
		case common.AoE3:
			game = age3.CreateGame()
		}
		GameTitles[gameTitle] = game
	}
}
