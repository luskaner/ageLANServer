package cmdUtils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/pelletier/go-toml/v2"
)

func WriteConfig(gameId string, config battleServerConfig.Config) (err error) {
	folder := battleServerConfig.Folder(gameId)
	err = os.MkdirAll(folder, 0755)
	if err != nil {
		return fmt.Errorf("error while creating folder \"%s\": %v (or it's parents)", battleServerConfig.Folder(gameId), err)
	}
	var tomlBytes []byte
	tomlBytes, err = toml.Marshal(config)
	if err != nil {
		err = fmt.Errorf("error while marshalling battle server config: %v", err)
	}
	var entries []fs.DirEntry
	entries, err = os.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("error while reading battle server config directory \"%s\": %v", folder, err)
	}
	i := -1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		index, localErr := battleServerConfig.ParseFileName(entry.Name())
		if localErr != nil {
			continue
		}
		i = index
	}
	name := battleServerConfig.Name(i + 1)
	err = os.WriteFile(filepath.Join(folder, name), tomlBytes, 0644)
	if err != nil {
		return fmt.Errorf("error while writing battle server config to file \"%s\": %v", name, err)
	}
	return
}
