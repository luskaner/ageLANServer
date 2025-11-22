package router

import (
	"net/http"
	"strings"

	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

var playAnonymousPaths = map[string]bool{
	"/Client/LoginWithSteam":                 true,
	"/MultiplayerServer/ListPartyQosServers": true,
	"/Event/WriteTelemetryEvents":            true,
}

func PlayfabMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !playAnonymousPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, playfab.StaticSuffix) {
			entityToken := playfab.Session(r)
			var exists bool
			if entityToken != "" {
				_, exists = playfab.Id(entityToken)
			}
			if !exists {
				shared.RespondError(
					&w,
					401,
					"Unauthorized",
					401,
					"Invalid X-EntityToken header",
					"",
				)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
