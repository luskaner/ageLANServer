package MultiplayerServer

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

func ListPartyQosServers(w http.ResponseWriter, _ *http.Request) {
	shared.RespondNotAvailable(&w)
}
