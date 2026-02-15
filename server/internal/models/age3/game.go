package age3

import (
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateMainGame(
		common.GameAoE3,
		&models.CreateMainGameOpts{
			BattleServer: &models.BattleServerOpts{
				Name:    "null",
				OobPort: true,
			},
		},
	)
}
