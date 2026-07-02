//go:build !windows

package game

func game(process string) (gameId string) {
	if steamProcessToGame == nil {
		steamProcessToGame = make(map[string]string, len(gameToSteamProcess))
		for curGameId, procName := range gameToSteamProcess {
			steamProcessToGame[procName] = curGameId
		}
	}
	return steamProcessToGame[process]
}
