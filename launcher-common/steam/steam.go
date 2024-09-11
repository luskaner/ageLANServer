package steam

import (
	"fmt"
	"github.com/andygrunwald/vdf"
	"os"
	"path"
)

const appID = "813780"

func OpenUri() string {
	return fmt.Sprintf("steam://rungameid/%s", appID)
}

func GameInstalled() bool {
	p := HomeDirPath()
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
	for _, folder := range libraryFolders {
		folderMap, ok = folder.(map[string]interface{})
		if !ok {
			continue
		}
		var apps map[string]interface{}
		apps, ok = folderMap["apps"].(map[string]interface{})
		if !ok {
			continue
		}
		if _, exists := apps[appID]; exists {
			return true
		}
	}
	return false
}
