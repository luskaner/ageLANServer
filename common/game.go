package common

import mapset "github.com/deckarep/golang-set/v2"

type GameTitle string

const (
	AoE1 GameTitle = "age1"
	AoE2 GameTitle = "age2"
	AoE3 GameTitle = "age3"
	// AoE4 Unsupported
	AoE4 GameTitle = "age4"
	// AoM Unsupported
	AoM GameTitle = "athens"
)

var SupportedGameTitleSlice = []GameTitle{AoE1, AoE2, AoE3}
var SupportedGameTitles = mapset.NewThreadUnsafeSet[GameTitle](SupportedGameTitleSlice...)
var AllGameTitles = SupportedGameTitles.Union(mapset.NewThreadUnsafeSet[GameTitle](AoE4, AoM))

func (g *GameTitle) Description() string {
	switch *g {
	case AoE1:
		return "Age of Empires: Definitive Edition"
	case AoE2:
		return "Age of Empires II: Definitive Edition"
	case AoE3:
		return "Age of Empires III: Definitive Edition"
	case AoE4:
		return "Age of Empires IV: Anniversary Edition"
	case AoM:
		return "Age of Mythology: Retold"
	default:
		return string(*g)
	}
}
