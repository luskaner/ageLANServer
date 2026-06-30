package game

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common/game"
)

var gameToSteamProcess map[string]string
var steamProcessToGame map[string]string

func init() {
	gameToSteamProcess = map[string]string{
		game.AoE1: "AoEDE_s.exe",
		game.AoE2: "AoE2DE_s.exe",
		game.AoE3: "AoE3DE_s.exe",
		game.AoE4: "RelicCardinal.exe",
		game.AoM:  "AoMRT_s.exe",
	}
}

func steamProcess(gameId string) string {
	return gameToSteamProcess[gameId]
}

func Processes(gameId string, steam bool, xbox bool) []string {
	processes := mapset.NewThreadUnsafeSet[string]()
	if steam {
		processes.Add(steamProcess(gameId))
	}
	if xbox {
		processes.Add(xboxProcess(gameId))
	}
	return processes.ToSlice()
}
