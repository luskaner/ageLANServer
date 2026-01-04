package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens"
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
			var sess *playfab.SessionData
			entityToken := r.Header.Get("X-Entitytoken")
			if entityToken != "" {
				var exists bool
				sessions := models.Gg[*athens.Game](r).PlayfabSessions
				if sess, exists = sessions.GetById(entityToken); exists {
					sessions.ResetExpiry(sess.EntityToken())
					ctx := context.WithValue(r.Context(), "session", sess)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
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
		next.ServeHTTP(w, r)
	})
}
