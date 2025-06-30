package account

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
)

func SetCrossplayEnabled(w http.ResponseWriter, r *http.Request) {
	// Crossplay is always enabled regardless of the value sent
	enable := "1"
	if models.G(r).Title() == common.AoE1 {
		enable = r.PostFormValue("crossplayEnabled")
	} else {
		enable = r.PostFormValue("enable")
	}
	if enable == "1" {
		i.JSON(&w, i.A{0})
	} else {
		// Do not accept disabling it
		i.JSON(&w, i.A{2})
	}
}
