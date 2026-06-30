package resolver

func resolveAutoPath(gameId string, battleServerPath string) (path string) {
	if path = steamPath(gameId, battleServerPath); path != "" {
		return
	}
	if path = steamCrossOverPath(gameId, battleServerPath); path != "" {
		return
	}
	return steamWinePath(gameId, battleServerPath)
}
