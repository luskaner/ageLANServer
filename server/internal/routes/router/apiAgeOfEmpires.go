package router

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/routes/apiAgeOfEmpires/textmoderation"
)

type ApiAgeOfEmpires struct {
	Proxy
}

func (a *ApiAgeOfEmpires) Name() string {
	return common.ApiAgeOfEmpires
}

func (a *ApiAgeOfEmpires) Initialize(gameId string) bool {
	return gameId == common.GameAoE3 || gameId == common.GameAoM
}

func (a *ApiAgeOfEmpires) InitializeRoutes(gameId string, next http.Handler) http.Handler {
	a.Proxy = NewProxy(common.ApiAgeOfEmpires, func(gameId string, next http.Handler) http.Handler {
		a.group.HandleFunc("POST", "/textmoderation", textmoderation.TextModeration)
		return a.group.mux
	})
	return a.Proxy.InitializeRoutes(gameId, next)
}
