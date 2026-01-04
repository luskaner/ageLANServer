package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/luskaner/ageLANServer/server/internal/models"
)

var sessAnonymousPaths = map[string]bool{
	"/game/msstore/getStoreTokens":      true,
	"/game/login/platformlogin":         true,
	"/wss/":                             true,
	"/game/news/getNews":                true,
	"/game/Challenge/getChallenges":     true,
	"/game/item/getItemBundleItemsJson": true,
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !sessAnonymousPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, "/cloudfiles/") {
			sessionID := r.FormValue("sessionID")
			if sessionID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			sessions := models.G(r).Sessions()
			sess, ok := sessions.GetById(sessionID)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			sessions.ResetExpiry(sess.Id())
			ctx := context.WithValue(r.Context(), "session", sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
