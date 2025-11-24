package models

import "net/http"

type Game interface {
	BattleServers() *MainBattleServers
	Resources() *MainResources
	Users() *MainUsers
	Advertisements() *MainAdvertisements
	ChatChannels() *MainChatChannels
	Title() string
}

func G(r *http.Request) Game {
	return Gg[Game](r)
}

func Gg[T Game](r *http.Request) T {
	return r.Context().Value("game").(T)
}
