package communityEvent

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
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
