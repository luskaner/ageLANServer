package age2

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateGame(common.GameAoE2, mapset.NewSet[string]("itemBundleItems.json", "itemDefinitions.json"))
}
