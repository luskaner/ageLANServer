package userData

import (
	"os"
	"path/filepath"
	"strconv"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

func profileFolder(gameId string) string {
	var p string
	switch gameId {
	case common.GameAoE1:
		p = "Users"
	}
	return p
}

func Profiles(gameId string) (err error, profiles mapset.Set[Data]) {
	var entries []os.DirEntry
	baseDir := filepath.Join(Path(gameId), profileFolder(gameId))
	entries, err = os.ReadDir(baseDir)
	if err != nil {
		return
	}
	profiles = mapset.NewThreadUnsafeSet[Data]()
	for _, entry := range entries {
		if entry.IsDir() {
			t, _ := typ(entry.Name())
			if gameId != common.GameAoE1 {
				if _, localErr := strconv.ParseUint(entry.Name(), 10, 64); localErr != nil {
					continue
				}
			}
			profiles.Add(Data{t, filepath.Join(baseDir, entry.Name())})
		}
	}
	return
}
