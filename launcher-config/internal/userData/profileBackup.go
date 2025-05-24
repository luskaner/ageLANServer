package userData

import (
	"github.com/luskaner/ageLANServer/common"
	"os"
	"strconv"
	"strings"
)

var profiles []Data

func setProfileData(staticBasePath string, gameId string) bool {
	profiles = make([]Data, 0)
	entries, err := os.ReadDir(path(staticBasePath, gameId))
	if err != nil {
		return false
	}
	var valid bool
	for _, entry := range entries {
		if entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".bak") || strings.HasSuffix(entry.Name(), ".lan") {
				valid = false
			} else if gameId == common.GameAoE1 {
				valid = true
			} else {
				_, err = strconv.ParseUint(entry.Name(), 10, 64)
				valid = err == nil
			}
			if valid {
				profiles = append(profiles, Data{entry.Name()})
			}
		}
	}
	return true
}

func runProfileMethod(staticBasePath string, gameId string, mainMethod func(staticBasePath string, gameId string, data Data) bool, cleanMethod func(staticBasePath string, gameId string, data Data) bool, stopOnFailed bool) bool {
	if !setProfileData(staticBasePath, gameId) {
		return false
	}
	for i := range profiles {
		if !mainMethod(staticBasePath, gameId, profiles[i]) {
			if !stopOnFailed {
				continue
			}
			for j := i - 1; j >= 0; j-- {
				_ = cleanMethod(staticBasePath, gameId, profiles[j])
			}
			return false
		}
	}
	return true
}

func backupProfile(staticBasePath string, gameId string, data Data) bool {
	return data.Backup(staticBasePath, gameId)
}

func restoreProfile(staticBasePath string, gameId string, data Data) bool {
	return data.Restore(staticBasePath, gameId)
}

func BackupProfiles(staticBasePath string, gameId string) bool {
	return runProfileMethod(staticBasePath, gameId, backupProfile, restoreProfile, true)
}

func RestoreProfiles(staticBasePath string, gameId string, reverseFailed bool) bool {
	return runProfileMethod(staticBasePath, gameId, restoreProfile, backupProfile, reverseFailed)
}
