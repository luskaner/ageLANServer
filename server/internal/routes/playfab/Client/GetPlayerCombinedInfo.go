package Client

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/middleware"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type userReadonlyData struct{}

type infoResultPayload struct {
	UserInventory           []any
	UserDataVersion         int
	userReadonlyData        `json:"UserReadOnlyData"`
	UserReadOnlyDataVersion int
	CharacterInventories    []any
}
type getPlayerCombinedInfoRequest struct {
	PlayFabId         string
	infoResultPayload `json:"InfoResultPayload"`
}

func GetPlayerCombinedInfo(w http.ResponseWriter, r *http.Request) {
	playfabId, _ := playfab.Id(middleware.PlayfabSession(r))
	shared.RespondOK(
		&w,
		getPlayerCombinedInfoRequest{
			PlayFabId: playfabId,
			infoResultPayload: infoResultPayload{
				UserInventory:        []any{},
				CharacterInventories: []any{},
			},
		},
	)
}
