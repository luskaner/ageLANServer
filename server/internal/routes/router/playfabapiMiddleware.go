package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client/shared"
)

var playAnonymousPaths = map[string]bool{
	"/Client/LoginWithSteam":                 true,
	"/Client/LoginWithCustomID":              true,
	"/MultiplayerServer/ListPartyQosServers": true,
	"/Event/WriteTelemetryEvents":            true,
}

func PlayfabMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !playAnonymousPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, playfab.StaticSuffix) {
			game := models.Gg[playfab.Game](r)
			var authHeader string
			if game.Title() == common.GameAoE4 {
				authHeader = "X-Sessionticket"
			} else {
				authHeader = "X-Entitytoken"
			}
			var sess *playfab.SessionData
			token := r.Header.Get(authHeader)
			if token != "" {
				var exists bool
				sessions := game.PlayfabSessions()
				if sess, exists = sessions.GetById(token); exists {
					sessions.ResetExpiry(sess.Token())
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
				fmt.Sprintf("Invalid %s header", authHeader),
				"",
			)
			return
		}
		next.ServeHTTP(w, r)
	})
}
