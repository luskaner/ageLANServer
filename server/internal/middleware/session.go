package middleware

import (
	"context"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"net/http"
	"strings"
)

var anonymousPaths = map[string]bool{
	"/test":                        true,
	"/game/msstore/getStoreTokens": true,
	"/game/login/platformlogin":    true,
	"/wss/":                        true,
	"/game/news/getNews":           true,
}

func Session(r *http.Request) *models.Session {
	sessAny, ok := r.Context().Value("session").(*models.Session)
	if !ok {
		panic("Session should have been set already")
	}
	return sessAny
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !anonymousPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, "/cloudfiles/") {
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
