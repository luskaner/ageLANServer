package userData

import (
	"github.com/luskaner/ageLANServer/common"
	"path/filepath"
)

func Metadata(gameTitle common.GameTitle) Data {
	var p string
	switch gameTitle {
	case common.AoE2:
		p = "metadata"
	case common.AoE3:
		p = filepath.Join("Common", "RLink")
	}
	return Data{Path: p}
}
