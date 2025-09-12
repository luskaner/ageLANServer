package userData

import (
	"path/filepath"

	"github.com/luskaner/ageLANServer/common"
)

func Metadata(gameId string) Data {
	var path string
	switch gameId {
	case common.GameAoE2:
		path = "metadata"
	case common.GameAoE3:
		path = filepath.Join("Common", "RLink")
	case common.GameAoM:
		path = filepath.Join("temp", "RLink")
	}
	return Data{Path: path}
}
