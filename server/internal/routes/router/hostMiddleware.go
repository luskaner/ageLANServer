package router

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
)

func HostMiddleware(gameId string) http.Handler {
	var playfabHandler http.Handler
	var apiAoEHandler http.Handler
	if gameId == common.GameAoM {
		playfab := &PlayfabApi{}
		playfabHandler = playfab.InitializeRoutes("", nil)
	}
	if gameId == common.GameAoE3 || gameId == common.GameAoM {
		aoeApi := &ApiAgeOfEmpires{}
		apiAoEHandler = aoeApi.InitializeRoutes("", nil)
	}
	game := Game{}
	gameHandler := game.InitializeRoutes(gameId, nil)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiAoEHandler != nil {
			if subdomain, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.ApiAgeOfEmpiresDomain && subdomain == common.ApiAgeOfEmpiresSubdomain {
				apiAoEHandler.ServeHTTP(w, r)
			}
		}
		if playfabHandler != nil {
			if _, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.PlayFabDomain {
				playfabHandler.ServeHTTP(w, r)
			}
		}
		gameHandler.ServeHTTP(w, r)
	})
}
