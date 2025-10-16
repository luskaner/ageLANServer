package router

import (
	"net/http"

	"github.com/luskaner/ageLANServer/server/internal/routes/apiAgeOfEmpires"
	"github.com/luskaner/ageLANServer/server/internal/routes/apiAgeOfEmpires/textmoderation"
)

type ApiAgeOfEmpires struct {
	Router
}

func (a *ApiAgeOfEmpires) Name() string {
	return "api.ageofempires"
}

func (a *ApiAgeOfEmpires) InitializeRoutes(_ string, _ http.Handler) http.Handler {
	a.initialize()
	a.group.HandleFunc("POST", "/textmoderation", textmoderation.TextModeration)
	if proxy := apiAgeOfEmpires.Root(); proxy != nil {
		a.group.HandlePath("/", proxy)
	}
	return a.group.mux
}
