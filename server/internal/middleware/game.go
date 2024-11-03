package middleware

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/leaderboard/age3"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"net/http"
)

var gamePathHandlers = map[string]map[string]map[string]http.HandlerFunc{
	"age3": {
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

func GameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ignoredPaths[r.URL.Path] {
			gameId := r.URL.Query().Get("title")
			if gameId == "" && r.Method == http.MethodPost {
				if err := r.ParseForm(); err == nil {
					gameId = r.Form.Get("title")
				}
			}
			if gameId == "" {
				session, ok := Session(r)
				if ok {
					gameId = session.GetGameId()
				}
			}
			gameSet := mapset.NewSet[string](viper.GetStringSlice("Games")...)
			if !gameSet.ContainsOne(gameId) {
				http.Error(w, "Unavailable game type", http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), "game", initializer.Games[gameId])
			req := r.WithContext(ctx)
			if gameHandlers, ok := gamePathHandlers[gameId]; ok {
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
