package chat

import (
	i "aoe2DELanServer/internal"
	"net/http"
)

func GetChatChannels(w http.ResponseWriter, _ *http.Request) {
	// TODO: What even are chat channels? plus the server seems to always return the same thing
	i.JSON(&w, i.A{0, i.A{}, 100})
}
