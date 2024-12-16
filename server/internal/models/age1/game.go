package age1

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

func CreateGame() models.Game {
	return models.CreateGame(common.GameAoE1, mapset.NewThreadUnsafeSet[string]("itemDefinitions.json"))
}
