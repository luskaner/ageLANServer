package steam

import (
	"fmt"
	"github.com/andygrunwald/vdf"
	"github.com/luskaner/ageLANServer/common"
	"os"
	"path"
	"path/filepath"
)

type Game struct {
	AppId string
}

func NewGame(id string) Game {
	return Game{AppId: AppId(id)}
}

func AppId(id string) string {
	switch id {
	case common.GameAoE1:
		return "1017900"
	case common.GameAoE2:
		return "813780"
	case common.GameAoE3:
		return "933110"
	default:
		return ""
	}
}

func (g Game) OpenUri() string {
	return fmt.Sprintf("steam://rungameid/%s", g.AppId)
}

func (g Game) GameInstalled() bool {
	p := ConfigPath()
	if p == "" {
		return false
	}
	f, err := os.Open(path.Join(p, "config", "libraryfolders.vdf"))
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Close()
	}()
	parser := vdf.NewParser(f)
	var data map[string]interface{}
	data, err = parser.Parse()
	if err != nil {
		return false
	}
	libraryFolders, ok := data["libraryfolders"].(map[string]interface{})
	if !ok {
		return false
	}
	var folderMap map[string]interface{}
	var stat os.FileInfo
	for _, folder := range libraryFolders {
		folderMap, ok = folder.(map[string]interface{})
		if !ok {
			continue
		}
		libraryPath := folderMap["path"].(string)
		if stat, err = os.Stat(filepath.Join(libraryPath, "steamapps", fmt.Sprintf("appmanifest_%s.acf", g.AppId))); err == nil && !stat.IsDir() {
			return true
		}
	}
	return false
}
