package age3

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateGame(common.GameAoE3, mapset.NewSet[string]("itemDefinitions.json"))
}
