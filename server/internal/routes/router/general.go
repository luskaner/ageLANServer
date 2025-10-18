package router

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/luskaner/ageLANServer/server/internal"
	cacertPem "github.com/luskaner/ageLANServer/server/internal/routes/cacert.pem"
	"github.com/luskaner/ageLANServer/server/internal/routes/test"
)

type General struct {
	Router
	Writer io.Writer
}

func (g *General) Name() string {
	return "general"
}

func (g *General) InitializeRoutes(gameId string, next http.Handler) http.Handler {
	g.initialize()
	writer := internal.NewPrefixedWriter(g.Writer, gameId, g.Name())
	g.group.Handle(
		"GET",
		"/test",
		handlers.LoggingHandler(writer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			test.Test(w, r)
		})),
	)
	g.group.Handle(
		"GET",
		"/cacert.pem",
		handlers.LoggingHandler(writer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cacertPem.CacertPem(w, r)
		})),
	)
	g.group.HandlePath("/", next)
	return g.group.mux
}
