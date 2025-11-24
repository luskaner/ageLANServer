package communityEvent

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens"
)

func GetAvailableCommunityEvents(w http.ResponseWriter, r *http.Request) {
	var response i.A
	game := models.G(r)
	if game.Title() == common.GameAoM {
		response = game.(*athens.Game).CommunityEventsEncoded()
	} else {
		response = i.A{0, i.A{}, i.A{}}
		if game.Title() == common.GameAoE2 {
			response = append(
				response,
				i.A{}, i.A{}, i.A{}, i.A{},
			)
		}
	}
	i.JSON(&w, response)
}
