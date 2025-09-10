package age2

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateGame(common.GameAoE2, mapset.NewThreadUnsafeSet[string]("itemBundleItems.json", "itemDefinitions.json"), true, "true")
}
