package common

import mapset "github.com/deckarep/golang-set/v2"

const (
	GameAoE2 = "age2"
	GameAoE3 = "age3"
)

var SupportedGames = mapset.NewSet[string](GameAoE2, GameAoE3)
