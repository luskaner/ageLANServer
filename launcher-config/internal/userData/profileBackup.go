package userData

import (
	commonUserData "github.com/luskaner/ageLANServer/launcher-common/userData"
)

var profiles []Data

func setProfileData(gameId string) bool {
	profiles = make([]Data, 0)
	err, commonProfiles := commonUserData.Profiles(gameId)
	if err != nil {
		return false
	}
	for entry := range commonProfiles.Iter() {
		if entry.Type == commonUserData.TypeActive {
			profiles = append(profiles, Data{entry.Path})
		}
	}
	return true
}

func runProfileMethod(gameId string, mainMethod func(data Data) bool, cleanMethod func(data Data) bool, stopOnFailed bool) bool {
	if !setProfileData(gameId) {
		return false
	}
	for i := range profiles {
		if !mainMethod(profiles[i]) {
			if !stopOnFailed {
				continue
			}
			for j := i - 1; j >= 0; j-- {
				_ = cleanMethod(profiles[j])
			}
			return false
		}
	}
	return true
}

func backupProfile(data Data) bool {
	return data.Backup()
}

func restoreProfile(data Data) bool {
	return data.Restore()
}

func BackupProfiles(gameId string) bool {
	return runProfileMethod(gameId, backupProfile, restoreProfile, true)
}

func RestoreProfiles(gameId string, reverseFailed bool) bool {
	return runProfileMethod(gameId, restoreProfile, backupProfile, reverseFailed)
}
