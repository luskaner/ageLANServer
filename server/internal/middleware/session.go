package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/luskaner/ageLANServer/server/internal/models"
)

var sessAnonymousPaths = map[string]bool{
	"/cacert.pem":                       true,
	"/test":                             true,
	"/game/msstore/getStoreTokens":      true,
	"/game/login/platformlogin":         true,
	"/wss/":                             true,
	"/game/news/getNews":                true,
	"/game/Challenge/getChallenges":     true,
	"/game/item/getItemBundleItemsJson": true,
}

func SessionOrPanic(r *http.Request) *models.Session {
	sessAny, ok := session(r)
	if !ok {
		panic("SessionOrPanic should have been set already")
	}
	return sessAny
}

func session(r *http.Request) (*models.Session, bool) {
	sess, ok := r.Context().Value("session").(*models.Session)
	return sess, ok
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !sessAnonymousPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, "/cloudfiles/") && !isPlayfab(r) && !isApiAgeOfEmpires(r) {
			sessionID := r.URL.Query().Get("sessionID")
			if sessionID == "" {
				err := r.ParseForm()
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				sessionID = r.Form.Get("sessionID")
			}
			sess, ok := models.GetSessionById(sessionID)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			sess.ResetExpiryTimer()
			ctx := context.WithValue(r.Context(), "session", sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
