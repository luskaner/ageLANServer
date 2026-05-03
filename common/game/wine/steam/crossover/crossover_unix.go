//go:build !windows

package crossover

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
)

const prefixEnvVar = "$CX_BOTTLE"

func Prefix(gameId string) string {
	var prefix string
	prefix = os.ExpandEnv(prefixEnvVar)
	if prefix == "" {
		prefix = defaultBottleName(gameId)
	}
	for _, baseDir := range baseDirs {
		if f, err := os.Stat(filepath.Join(os.ExpandEnv(baseDir), prefix)); err == nil && f.IsDir() {
			return prefix
		}
	}
	return ""
}

func defaultBottleName(gameId string) (name string) {
	switch gameId {
	case common.GameAoE1:
		name = "Age_of_Empires_Definitive_Edition"
	case common.GameAoE2:
		name = "Age_of_Empires_II_Definitive_Edition"
	case common.GameAoE3:
		name = "Age_of_Empires_III_Definitive_Edition"
	case common.GameAoE4:
		name = "Age_of_Empires_IV_Anniversary_Edition"
	case common.GameAoM:
		name = "Age_of_Mythology_Retold"
	}
	return
}
