package router

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
)

func HostMiddleware(gameId string, writer io.Writer) http.Handler {
	writers := make(map[*http.Handler]*internal.PrefixedWriter)
	var playfabHandler http.Handler
	var apiAoEHandler http.Handler
	if gameId == common.GameAoM {
		playfab := &PlayfabApi{}
		playfabHandler = playfab.InitializeRoutes("", nil)
		writers[&playfabHandler] = internal.NewPrefixedWriter(writer, gameId, playfab.Name())
	}
	if gameId == common.GameAoE3 || gameId == common.GameAoM {
		aoeApi := &ApiAgeOfEmpires{}
		apiAoEHandler = aoeApi.InitializeRoutes("", nil)
		writers[&apiAoEHandler] = internal.NewPrefixedWriter(writer, gameId, aoeApi.Name())
	}
	game := Game{}
	gameHandler := game.InitializeRoutes(gameId, nil)
	writers[&gameHandler] = internal.NewPrefixedWriter(writer, gameId, game.Name())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var next *http.Handler
		if apiAoEHandler != nil {
			if subdomain, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.ApiAgeOfEmpiresDomain && subdomain == common.ApiAgeOfEmpiresSubdomain {
				next = &apiAoEHandler
			}
		}
		if next == nil && playfabHandler != nil {
			if _, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.PlayFabDomain {
				next = &playfabHandler
			}
		}
		if next == nil {
			next = &gameHandler
		}
		handlers.LoggingHandler(writers[next], *next).ServeHTTP(w, r)
	})
}
