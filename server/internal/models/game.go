package models

import (
	"github.com/luskaner/ageLANServer/common"
	"net/http"
)

type Game interface {
	Resources() *MainResources
	Users() *MainUsers
	Advertisements() *MainAdvertisements
	ChatChannels() *MainChatChannels
	Title() common.GameTitle
}

func G(r *http.Request) Game {
	return r.Context().Value("game").(Game)
}
