package crossover

import (
	"strings"
)

var baseDirs = []string{"$HOME/.cxoffice"}

func defaultBottleName(gameId string) (name string) {
	return strings.ReplaceAll(baseDefaultBottleName(gameId), " ", "_")
}
