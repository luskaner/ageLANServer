package chat

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetChatChannels(w http.ResponseWriter, _ *http.Request) {
	// What even are chat channels? plus the server seems to always return the same thing
	i.JSON(&w, i.A{0, i.A{}, 100})
}
