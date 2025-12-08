package models

import "net/http"

type Game interface {
	BattleServers() BattleServers
	Resources() Resources
	Users() Users
	Advertisements() Advertisements
	ChatChannels() ChatChannels
	Sessions() Sessions
	Title() string
}

func G(r *http.Request) Game {
	return Gg[Game](r)
}

func Gg[T Game](r *http.Request) T {
	return r.Context().Value("game").(T)
}
