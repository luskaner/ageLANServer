package leaderboard

import (
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/leaderboard/shared"
	"net/http"
)

func GetStatGroupsByProfileIDs(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(
		r,
		r.URL.Query().Get("profileids"),
		true,
		models.G(r).Title() != common.GameAoE3,
	)
	i.JSON(&w, response)
}
