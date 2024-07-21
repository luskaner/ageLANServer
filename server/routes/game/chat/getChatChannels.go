package chat

import (
	"net/http"
	i "server/internal"
)

func GetChatChannels(w http.ResponseWriter, _ *http.Request) {
	// What even are chat channels? plus the server seems to always return the same thing
	i.JSON(&w, i.A{0, i.A{}, 100})
}
