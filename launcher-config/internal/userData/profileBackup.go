package userData

import (
	"os"
	"strconv"
	"strings"

	"github.com/luskaner/ageLANServer/common"
)

var profiles []Data

func setProfileData(gameId string) bool {
	profiles = make([]Data, 0)
	entries, err := os.ReadDir(path(gameId))
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

func runProfileMethod(gameId string, mainMethod func(gameId string, data Data) bool, cleanMethod func(gameId string, data Data) bool, stopOnFailed bool) bool {
	if !setProfileData(gameId) {
		return false
	}
	for i := range profiles {
		if !mainMethod(gameId, profiles[i]) {
			if !stopOnFailed {
				continue
			}
			for j := i - 1; j >= 0; j-- {
				_ = cleanMethod(gameId, profiles[j])
			}
			return false
		}
	}
	return true
}

func backupProfile(gameId string, data Data) bool {
	return data.Backup(gameId)
}

func restoreProfile(gameId string, data Data) bool {
	return data.Restore(gameId)
}

func BackupProfiles(gameId string) bool {
	return runProfileMethod(gameId, backupProfile, restoreProfile, true)
}

func RestoreProfiles(gameId string, reverseFailed bool) bool {
	return runProfileMethod(gameId, restoreProfile, backupProfile, reverseFailed)
}
