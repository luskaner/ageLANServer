package cmdUtils

import (
	"os"
	"path/filepath"

	"battle-server-manager/internal"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
)

func Kill(config battleServer.Config) bool {
	proc, err := process.FindProcess(int(config.PID))
	if err == nil && proc != nil {
		str := "\t\tProcess still running, killing it..."
		if err = process.KillProc(proc); err == nil {
			commonLogger.Println(str + " OK")
			return true
		}
		commonLogger.Println(str+" failed with error: ", err)
		return false
	}
	return true
}

func remove(gameId string, config battleServer.Config) bool {
	commonLogger.Println("\tRemoving:", config.Region)
	_ = Kill(config)
	folder := battleServer.Folder(gameId)
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

// RemoveGames resolves the given game ids and removes their battle server
// configs, logging progress. When onlyInvalid is true only configs that fail
// validation are removed.
func RemoveGames(gameIds *[]string, onlyInvalid bool) (err error, exitCode int) {
	var games mapset.Set[string]
	games, err = ParsedGameIds(gameIds)
	if err != nil {
		commonLogger.Println(err.Error())
		exitCode = internal.ErrGames
		return
	}
	var configs []battleServer.Config
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		configs, err = battleServer.Configs(g, false)
		if err != nil {
			commonLogger.Printf("\t%s\n", err)
			continue
		}
		if !Remove(g, configs, onlyInvalid) {
			commonLogger.Println("\tNo configuration needs it.")
		}
	}
	return
}

func Remove(gameId string, configs []battleServer.Config, onlyInvalid bool) bool {
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
