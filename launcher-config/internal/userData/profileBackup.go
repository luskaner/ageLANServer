package userData

import (
	"os"
	"strconv"
)

var profiles []Data

func setProfileData() bool {
	profiles = make([]Data, 0)
	entries, err := os.ReadDir(Path())
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

func runProfileMethod(mainMethod func(data Data) bool, cleanMethod func(data Data) bool, stopOnFailed bool) bool {
	if !setProfileData() {
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

func BackupProfiles() bool {
	return runProfileMethod(backupProfile, restoreProfile, true)
}

func RestoreProfiles(reverseFailed bool) bool {
	return runProfileMethod(restoreProfile, backupProfile, reverseFailed)
}
