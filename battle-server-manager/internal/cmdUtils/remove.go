package cmdUtils

import (
	"os"
	"path/filepath"

	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
)

func Kill(config battleServerConfig.Config) bool {
	proc, err := process.FindProcess(int(config.PID))
	if err == nil && proc != nil {
		str := "\t\tProcess still running, killing it..."
		if err = process.KillProc(proc); err == nil {
			commonLogger.Println(str + " OK")
			return true
		} else {
			commonLogger.Println(str+" failed with error: ", err)
			return false
		}
	}
	return true
}

func remove(gameId string, config battleServerConfig.Config) bool {
	commonLogger.Println("\tRemoving:", config.Region)
	_ = Kill(config)
	folder := battleServerConfig.Folder(gameId)
	if f, err := os.Stat(folder); err != nil || !f.IsDir() {
		return false
	}
	fullPath := filepath.Join(folder, config.Path())
	if f, err := os.Stat(fullPath); err == nil && !f.IsDir() {
		str := "\t\tRemoving config file..."
		if err := os.Remove(fullPath); err == nil {
			commonLogger.Println(str + " OK")
		} else {
			commonLogger.Println(str+" failed with error: ", err)
		}
	} else {
		commonLogger.Println("Failed with error: ", err)
	}
	return true
}

func Remove(gameId string, configs []battleServerConfig.Config, onlyInvalid bool) bool {
	var removedAny bool
	for _, config := range configs {
		var doRemove bool
		if onlyInvalid {
			if !config.Validate() {
				doRemove = true
			}
		} else {
			doRemove = true
		}
		if doRemove {
			removed := remove(gameId, config)
			removedAny = removedAny || removed
		}
	}
	return removedAny
}
