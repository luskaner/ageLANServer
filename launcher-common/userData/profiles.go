package userData

import (
	"os"
	"path/filepath"
	"strconv"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
)

func (u *Path) profileFolder() string {
	var p string
	switch u.GameId() {
	case common.GameAoE1, common.GameAoE4:
		p = "Users"
	}
	return p
}

func (u *Path) Profiles() (err error, profiles mapset.Set[Data]) {
	var entries []os.DirEntry
	baseDir := filepath.Join(u.String(), u.profileFolder())
	entries, err = os.ReadDir(baseDir)
	if err != nil {
		return
	}
	profiles = mapset.NewThreadUnsafeSet[Data]()
	for _, entry := range entries {
		if entry.IsDir() {
			t, _ := typ(entry.Name())
			if u.gameId != common.GameAoE1 {
				if _, localErr := strconv.ParseUint(entry.Name(), 10, 64); localErr != nil {
					continue
				}
			}
			profiles.Add(Data{t, filepath.Join(baseDir, entry.Name())})
		}
	}
	return
}
