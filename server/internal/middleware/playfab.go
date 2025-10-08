package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

func isPlayfab(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, playfab.Prefix)
}

var playAnonymousPaths = map[string]bool{
	"/Client/LoginWithSteam":                 true,
	"/MultiplayerServer/ListPartyQosServers": true,
	"/Event/WriteTelemetryEvents":            true,
}

func PlayfabSession(r *http.Request) string {
	return r.Header.Get("X-Entitytoken")
}

func PlayfabMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.PlayFabDomain {
			if !playAnonymousPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, playfab.StaticSuffix+"/") {
				entityToken := PlayfabSession(r)
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
			r.URL.Path = path.Join(playfab.Prefix, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
