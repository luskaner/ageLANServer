package middleware

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/leaderboard/age3"
	"net/http"
	"strings"
)

var gamePathHandlers = map[common.GameTitle]map[string]map[string]http.HandlerFunc{
	common.AoE3: {
		"/game/leaderboard/setAvatarStatValues": {
			http.MethodPost: age3.SetAvatarStatValues,
		},
	},
}

var ignoredPaths = map[string]bool{
	"/":                            true,
	"/test":                        true,
	"/game/msstore/getStoreTokens": true,
	"/wss/":                        true,
	"/game/news/getNews":           true,
}

func GameMiddleware(gameSet mapset.Set[common.GameTitle], next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ignoredPaths[r.URL.Path] {
			game := r.URL.Query().Get("title")
			if game == "" && r.Method == http.MethodPost {
				if err := r.ParseForm(); err == nil {
					game = r.Form.Get("title")
				}
			}
			if game == "" && !strings.HasPrefix(r.URL.Path, "/cloudfiles/game/") {
				session := Session(r)
				game = session.Getgame()
			}
			if !gameSet.ContainsOne(common.GameTitle(game)) {
				http.Error(w, "Unavailable game type", http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), "game", initializer.GameTitles[common.GameTitle(game)])
			req := r.WithContext(ctx)
			if gameHandlers, ok := gamePathHandlers[common.GameTitle(game)]; ok {
				var pathHandlers map[string]http.HandlerFunc
				if pathHandlers, ok = gameHandlers[r.URL.Path]; ok {
					var handler http.HandlerFunc
					if handler, ok = pathHandlers[r.Method]; ok {
						handler.ServeHTTP(w, req)
						return
					}
				}
			}
			next.ServeHTTP(w, req)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
