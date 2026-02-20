package router

import (
	"context"
	"net/http"
	"time"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/login"
)

func LoginUserMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now().UTC()
		var req login.PlatformLoginRequest
		if err := i.Bind(r, &req); err != nil {
			login.PlatformLoginError(t, w)
			return
		}
		game := models.G(r)
		title := game.Title()
		var avatarStatDefinitions models.AvatarStatDefinitions = nil
		if title != common.GameAoE1 {
			avatarStatDefinitions = game.LeaderboardDefinitions().AvatarStatDefinitions()
		}
		u := game.Users().GetOrCreateUser(
			title,
			game.Items(),
			avatarStatDefinitions,
			r.RemoteAddr,
			req.MacAddress,
			req.AccountType == "XBOXLIVE",
			req.PlatformUserId,
			req.Alias,
		)
		ctx := context.WithValue(r.Context(), "user", u)
		ctx = context.WithValue(ctx, "request", req)
		ctx = context.WithValue(ctx, "time", t)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
