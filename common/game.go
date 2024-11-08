package common

import mapset "github.com/deckarep/golang-set/v2"

const (
	GameAoE1 = "age1"
	GameAoE2 = "age2"
	GameAoE3 = "age3"
)

var SupportedGames = mapset.NewSet[string](GameAoE1, GameAoE2, GameAoE3)
