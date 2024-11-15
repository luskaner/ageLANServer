package advertisement

import (
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strings"
)

type query struct {
	RelayRegions string `schema:"lanServerGuids"`
}

func GetLanAdvertisements(w http.ResponseWriter, r *http.Request) {
	var q query
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}, i.A{}})
		return
	}
	lanServerGuids := strings.Split(strings.Trim(q.RelayRegions, `[]"`), ",")
	lanServerGuidsMap := make(map[string]struct{}, len(lanServerGuids))
	for _, guid := range lanServerGuids {
		lanServerGuidsMap[guid] = struct{}{}
	}
	matches := func(adv *models.MainAdvertisement) bool {
		_, relayRegionMatches := lanServerGuidsMap[adv.GetRelayRegion()]
		return relayRegionMatches
	}
	findAdvertisements(w, r, models.G(r), matches)
}
