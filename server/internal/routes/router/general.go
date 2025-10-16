package router

import (
	"net/http"

	cacertPem "github.com/luskaner/ageLANServer/server/internal/routes/cacert.pem"
	"github.com/luskaner/ageLANServer/server/internal/routes/test"
)

type General struct {
	Router
}

func (g *General) Name() string {
	return "general"
}

func (g *General) InitializeRoutes(_ string, next http.Handler) http.Handler {
	g.initialize()
	g.group.HandleFunc("GET", "/test", test.Test)
	g.group.HandleFunc("GET", "/cacert.pem", cacertPem.CacertPem)
	g.group.HandlePath("/", next)
	return g.group.mux
}
