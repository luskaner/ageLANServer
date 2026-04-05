package router

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal/routes/apiAgeOfEmpires/textmoderation"
)

type ApiAgeOfEmpires struct {
	Proxy
	apiName string
}

func (a *ApiAgeOfEmpires) Name() string {
	return a.apiName
}

func (a *ApiAgeOfEmpires) Initialize(_ string) bool {
	return true
}

func (a *ApiAgeOfEmpires) InitializeRoutes(gameId string, next http.Handler) http.Handler {
	a.Proxy = NewProxy(a.Name(), func(_ string, _ http.Handler) http.Handler {
		a.group.HandleFunc("POST", "/textmoderation", textmoderation.TextModeration)
		return a.group.mux
	})
	return a.Proxy.InitializeRoutes(gameId, next)
}

func NewApiAgeOfEmpires() *ApiAgeOfEmpires {
	return &ApiAgeOfEmpires{
		apiName: common.ApiAgeOfEmpires,
	}
}

type Aoe4ApiAgeOfEmpires struct {
	ApiAgeOfEmpires
}

func (a *Aoe4ApiAgeOfEmpires) Initialize(gameId string) bool {
	return gameId == common.GameAoE4
}

func NewAoe4ApiAgeOfEmpires() *Aoe4ApiAgeOfEmpires {
	return &Aoe4ApiAgeOfEmpires{
		ApiAgeOfEmpires{
			apiName: common.Aoe4ApiAgeOfEmpires,
		},
	}
}
