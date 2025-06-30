package age3

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateGame(
		common.AoE3,
		mapset.NewThreadUnsafeSet[string]("itemDefinitions.json"),
		mapset.NewThreadUnsafeSet[string]("achievements.json"),
	)
}
