package router

import (
	"context"
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
)

func GameMiddleware(gameId string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "game", initializer.Games[gameId])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
