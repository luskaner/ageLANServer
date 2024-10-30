package models

import "net/http"

type Game interface {
	Resources() *MainResources
	Users() *MainUsers
	Advertisements() *MainAdvertisements
	ChatChannels() *MainChatChannels
	Title() string
}

func G(r *http.Request) Game {
	return r.Context().Value("game").(Game)
}
