package Client

import (
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

type getUserDataResponseDataInner struct {
	Value       string
	LastUpdated string
	Permission  string
}

type getUserDataResponseData struct {
	getUserDataResponseDataInner `json:"RLinkProfileID"`
}

type getUserDataResponse struct {
	getUserDataResponseData `json:"Data"`
	DataVersion             int
}

func GetUserData(w http.ResponseWriter, _ *http.Request) {
	shared.RespondOK(&w, getUserDataResponse{
		getUserDataResponseData: getUserDataResponseData{
			getUserDataResponseDataInner: getUserDataResponseDataInner{
				// At this point we don't know the user ID, return a non existing ID so the user is forced to update to the actual one.
				Value:       "0",
				LastUpdated: time.Now().Format("2006-01-02T15:04:05.999"),
				Permission:  "Public",
			},
		},
		// Assume it's the first call
		DataVersion: 1,
	})
}
