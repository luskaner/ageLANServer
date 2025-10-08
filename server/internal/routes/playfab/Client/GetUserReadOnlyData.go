package Client

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type data struct{}

type getUserReadOnlyDataResponse struct {
	DataVersion int32
	data        `json:"Data"`
}

func GetUserReadOnlyData(w http.ResponseWriter, _ *http.Request) {
	shared.RespondOK(&w, getUserReadOnlyDataResponse{})
}
