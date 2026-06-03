//go:build !windows

package game

func xboxProcess(_ string) string {
	// Should not be called
	return ""
}

func Game(process string) (gameId string) {
	if steamProcessToGame == nil {
		steamProcessToGame = make(map[string]string, len(gameToSteamProcess))
		for curGameId, procName := range gameToSteamProcess {
			steamProcessToGame[procName] = curGameId
		}
	}
	return steamProcessToGame[process]
}
