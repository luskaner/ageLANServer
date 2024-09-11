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

func openVdf(path string) (data map[string]interface{}, err error) {
	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	parser := vdf.NewParser(f)
	data, err = parser.Parse()
	return
}

func GameInstalled() bool {
	p := HomeDirPath()
	if p == "" {
		return false
	}
	data, err := openVdf(path.Join(p, "config", "libraryfolders.vdf"))
	if err != nil {
		return false
	}
	libraryFolders, ok := data["libraryfolders"].(map[string]interface{})
	if !ok {
		return false
	}

	for _, folder := range libraryFolders {
		var folderMap map[string]interface{}
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
