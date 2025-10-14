package Client

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type responseUpdateUserData struct {
	DataVersion int
}

func UpdateUserData(w http.ResponseWriter, _ *http.Request) {
	// Assume it's the second version
	shared.RespondOK(&w, responseUpdateUserData{DataVersion: 2})
}
