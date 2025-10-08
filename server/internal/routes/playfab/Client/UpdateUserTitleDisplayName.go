package Client

import (
	"encoding/json"
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type updateUserTitleDisplayNameRequest struct {
	DisplayName string
}

type updateUserTitleDisplayNameResponse struct {
	DisplayName string
}

func UpdateUserTitleDisplayName(w http.ResponseWriter, r *http.Request) {
	var req updateUserTitleDisplayNameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		shared.RespondBadRequest(&w)
		return
	}
	shared.RespondOK(
		&w,
		updateUserTitleDisplayNameResponse{
			DisplayName: req.DisplayName,
		},
	)
}
