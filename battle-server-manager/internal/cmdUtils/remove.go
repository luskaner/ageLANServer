package cmdUtils

import (
	"battle-server-manager/internal"
	"os"
	"path/filepath"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/process"
)

var (
	parsedGameIdsFnRemoveAll   = ParsedGameIds
	battleServerConfigsFn      = battleServer.Configs
	removeFnRemoveAll          = Remove
	findProcessFn              = process.FindProcess
	killProcFn                 = process.KillProc
)

func Kill(config battleServer.Config) bool {
	proc, err := findProcessFn(int(config.PID))
	if err == nil && proc != nil {
		str := "\t\tProcess still running, killing it..."
		if err = killProcFn(proc); err == nil {
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
	if f, err := os.Stat(fullPath); err != nil || f.IsDir() {
		commonLogger.Println("Failed with error: ", err)
		return false
	}
	str := "\t\tRemoving config file..."
	if err := os.Remove(fullPath); err != nil {
		commonLogger.Println(str+" failed with error: ", err)
		return false
	}
	commonLogger.Println(str + " OK")
	return true
}

func Remove(gameId string, configs []battleServer.Config, onlyInvalid bool) bool {
	var removedAny bool
	for _, config := range configs {
		var doRemove bool
		if onlyInvalid {
			if !config.Validate(false) {
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

func RemoveAll(onlyInvalid bool) (err error, exitCode int) {
	var games mapset.Set[string]
	games, err = parsedGameIdsFnRemoveAll(nil)
	if err != nil {
		commonLogger.Println(err.Error())
		exitCode = internal.ErrGames
		return
	}
	var configs []battleServer.Config
	for g := range games.Iter() {
		commonLogger.Printf("Game: %s\n", g)
		configs, err = battleServerConfigsFn(g, false, false)
		if err != nil {
			commonLogger.Printf("\t%s\n", err)
			continue
		}
		if !removeFnRemoveAll(g, configs, onlyInvalid) {
			commonLogger.Println("\tNo configuration needs it.")
		}
	}
	return
}
