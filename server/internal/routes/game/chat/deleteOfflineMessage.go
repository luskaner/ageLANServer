package chat

import (
	"net/http"

	i "github.com/luskaner/ageLANServer/server/internal"
)

func DeleteOfflineMessage(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{0})
}
