package userData

import (
	"github.com/luskaner/aoe2DELanServer/common"
	"path/filepath"
)

func Metadata(gameId string) Data {
	var path string
	switch gameId {
	case common.GameAoE2:
		path = "metadata"
	case common.GameAoE3:
		path = filepath.Join("Common", "RLink")
	}
	return Data{Path: path}
}
