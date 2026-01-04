package models

import "net/http"

type Game interface {
	Title() string
	Resources() Resources
	LeaderboardDefinitions() LeaderboardDefinitions
	BattleServers() BattleServers
	Users() Users
	Advertisements() Advertisements
	ChatChannels() ChatChannels
	Sessions() Sessions
}

func G(r *http.Request) Game {
	return Gg[Game](r)
}

func Gg[T Game](r *http.Request) T {
	return r.Context().Value("game").(T)
}
