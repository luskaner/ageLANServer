package userData

import (
	"os"
	"strconv"
)

var profiles []Data

func BackupProfiles() bool {
	entries, err := os.ReadDir(Path())
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			_, err := strconv.ParseUint(entry.Name(), 10, 64)
			if err == nil {
				profiles = append(profiles, Data{
					Path:     entry.Name(),
					Original: true,
				})
			}
		}
	}

	for i := range profiles {
		if !profiles[i].Backup() {
			return false
		}
	}
	return true
}

func RestoreProfiles() bool {
	for i := range profiles {
		if !profiles[i].Restore() {
			return false
		}
	}
	return true
}
