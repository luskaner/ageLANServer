package age1

import (
	"github.com/luskaner/ageLANServer/common/game"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateMainGame(
		game.AoE1,
		&models.CreateMainGameOpts{
			BattleServer: &models.BattleServerOpts{
				Name: "omit",
			},
		},
	)
}
