package age4

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
	commonPlayfab "github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

func CreateGame() models.Game {
	mainGame := models.CreateMainGame(
		common.GameAoE4,
		&models.CreateMainGameOpts{
			Resources: &models.ResourcesOpts{
				KeyedFilenames: mapset.NewThreadUnsafeSet[string](
					"itemBundleItems.json",
					"itemDefinitions.json",
					"levelRewardsTable.json",
				),
			},
			BattleServer: &models.BattleServerOpts{
				OobPort: true,
				Name:    "null",
			},
		},
	)
	g := &commonPlayfab.BaseGame{
		Game: mainGame,
	}
	g.PlayfabSessions().Initialize()
	return g
}
