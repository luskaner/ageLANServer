package chat

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func GetChatChannels(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, i.A{0, models.G(r).ChatChannels().Encode(), 100})
}
