package communityEvent

import (
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetAvailableCommunityEvents(w http.ResponseWriter, r *http.Request) {
	if models.G(r).Title() == common.GameAoE3 {
		i.JSON(&w, i.A{0, i.A{}, i.A{}})
	} else {
		i.JSON(&w, i.A{0, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}, i.A{}})
	}
}
