package age2

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateMainGame(
		common.GameAoE2,
		&models.CreateMainGameOpts{
			Resources: &models.ResourcesOpts{
				KeyedFilenames: mapset.NewThreadUnsafeSet[string]("itemBundleItems.json", "itemDefinitions.json"),
			},
		},
	)
}
