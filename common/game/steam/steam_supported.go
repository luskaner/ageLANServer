//go:build windows || linux

package steam

func NewGame(gameId string) (game *Game, ok bool) {
	return NewCustomGame(gameId, ConfigPath, ConfigPathAlt, func(s string) string {
		return s
	})
}
