package account

import (
	"net/http"
	i "server/internal"
)

func SetAvatarMetadata(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, i.A{}})
}
