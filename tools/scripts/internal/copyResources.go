package internal

import (
	"fmt"
	"log"
	"path/filepath"
	c "scripts/internal/constants"

	"github.com/luskaner/ageLANServer/common"
	p "github.com/luskaner/ageLANServer/common/paths"
)

func BuildResourcePath(module string) string {
	return filepath.Join(c.BuildDir, module, p.ResourcesDir)
}

func ResourcePath(module string) string {
	return filepath.Join(module, p.ResourcesDir)
}

func CopyMainConfig(module string) {
	if err := Cp(filepath.Join(ResourcePath(module), c.ConfigFileName), filepath.Join(BuildResourcePath(module), c.ConfigFileName)); err != nil {
		log.Fatal(err)
	}
}

func CopyGameConfigs(module string) {
	dstPath := BuildResourcePath(module)
	MkdirP(dstPath)
	src := filepath.Join(ResourcePath(module), fmt.Sprintf(c.GameConfigFileName, "game"))
	for game := range common.SupportedGames.Iter() {
		dst := filepath.Join(dstPath, fmt.Sprintf(c.GameConfigFileName, game))
		if err := Cp(src, dst); err != nil {
			log.Fatal(err)
		}
	}
}
