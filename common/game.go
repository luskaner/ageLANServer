package common

import mapset "github.com/deckarep/golang-set/v2"

const (
	GameAoE1 = "age1"
	GameAoE2 = "age2"
	GameAoE3 = "age3"
	GameAoE4 = "age4"
	GameAoM  = "athens"
)

var SupportedGames = mapset.NewThreadUnsafeSet[string](GameAoE1, GameAoE2, GameAoE3, GameAoE4, GameAoM)
var AllGames = SupportedGames
