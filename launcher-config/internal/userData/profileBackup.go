package userData

import (
	"os"
	"strconv"
)

var profiles []Data

func setProfileData(gameId string) bool {
	profiles = make([]Data, 0)
	entries, err := os.ReadDir(Path(gameId))
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			_, err = strconv.ParseUint(entry.Name(), 10, 64)
			if err == nil {
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
