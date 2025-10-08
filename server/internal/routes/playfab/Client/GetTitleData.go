package Client

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

func GetTitleData(w http.ResponseWriter, _ *http.Request) {
	shared.RespondNotAvailable(&w)
}
