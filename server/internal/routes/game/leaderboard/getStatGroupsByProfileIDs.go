package leaderboard

import (
	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/leaderboard/shared"
	"net/http"
)

func GetStatGroupsByProfileIDs(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(
		r,
		r.URL.Query().Get("profileids"),
		true,
		models.G(r).Title() != common.AoE3,
	)
	i.JSON(&w, response)
}
