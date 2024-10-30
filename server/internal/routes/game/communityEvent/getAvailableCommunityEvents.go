package communityEvent

import (
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetAvailableCommunityEvents(w http.ResponseWriter, r *http.Request) {
	response := i.A{0, i.A{}, i.A{}}
	if models.G(r).Title() == common.GameAoE2 {
		response = append(
			response,
			i.A{}, i.A{}, i.A{}, i.A{},
		)
	}
	i.JSON(&w, response)
}
