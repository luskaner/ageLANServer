package account

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"net/http"
)

func SetAvatarMetadata(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, i.A{2, i.A{}})
}
