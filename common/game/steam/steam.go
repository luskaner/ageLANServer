package steam

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andygrunwald/vdf"
	"github.com/luskaner/ageLANServer/common"
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
	case common.GameAoM:
		return "1934680"
	default:
		return ""
	}
}

func (g Game) OpenUri() string {
	return fmt.Sprintf("steam://rungameid/%s", g.AppId)
}

func (g Game) LibraryFolder() (folder string) {
	p := ConfigPath()
	if p == "" {
		return
	}
	f, err := os.Open(filepath.Join(p, "config", "libraryfolders.vdf"))
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	parser := vdf.NewParser(f)
	var data map[string]interface{}
	data, err = parser.Parse()
	if err != nil {
		return
	}
	libraryFolders, ok := data["libraryfolders"].(map[string]interface{})
	if !ok {
		return
	}
	var folderMap map[string]interface{}
	for _, dir := range libraryFolders {
		folderMap, ok = dir.(map[string]interface{})
		if !ok {
			continue
		}
		var apps map[string]interface{}
		apps, ok = folderMap["apps"].(map[string]interface{})
		if !ok {
			continue
		}
		if _, exists := apps[g.AppId]; exists {
			return folderMap["path"].(string)
		}
	}
	return
}

func (g Game) Path(libraryFolder string) (folder string) {
	basePath := filepath.Join(libraryFolder, "steamapps")
	f, err := os.Open(filepath.Join(basePath, fmt.Sprintf("appmanifest_%s.acf", g.AppId)))
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	parser := vdf.NewParser(f)
	var data map[string]any
	data, err = parser.Parse()
	if err != nil {
		return
	}
	folder = filepath.Join(basePath, "common", data["AppState"].(map[string]any)["installdir"].(string))
	if f, err := os.Stat(folder); err != nil || !f.IsDir() {
		return ""
	}
	return
}
