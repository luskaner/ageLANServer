package router

import (
	"net/http"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models/playfab"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Catalog"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Client"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/CloudScript"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Event"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Inventory"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/MultiplayerServer"
	"github.com/luskaner/ageLANServer/server/internal/routes/playfab/Party"
)

type PlayfabApi struct {
	Router
}

func (p *PlayfabApi) Name() string {
	return "playfabapi"
}

func (p *PlayfabApi) Check(r *http.Request) bool {
	if _, mainDomain, tld, err := internal.SplitDomain(r.Host); err == nil && tld == common.Tld && mainDomain == common.PlayFabDomain {
		return true
	}
	return false
}

func (p *PlayfabApi) Initialize(gameId string) bool {
	return gameId == common.GameAoM
}

func (p *PlayfabApi) InitializeRoutes(_ string, _ http.Handler) http.Handler {
	p.initialize()
	playfabClientGroup := p.group.Subgroup("/Client")
	playfabClientGroup.HandleFunc("POST", "/GetPlayerCombinedInfo", Client.GetPlayerCombinedInfo)
	playfabClientGroup.HandleFunc("POST", "/GetTime", Client.GetTime)
	playfabClientGroup.HandleFunc("POST", "/GetTitleData", Client.GetTitleData)
	playfabClientGroup.HandleFunc("POST", "/GetUserReadOnlyData", Client.GetUserReadOnlyData)
	playfabClientGroup.HandleFunc("POST", "/LoginWithSteam", Client.LoginWithSteam)
	playfabClientGroup.HandleFunc("POST", "/UpdateUserTitleDisplayName", Client.UpdateUserTitleDisplayName)

	playfabEventGroup := p.group.Subgroup("/Event")
	playfabEventGroup.HandleFunc("POST", "/WriteTelemetryEvents", Event.WriteTelemetryEvents)

	playfabInventoryGroup := p.group.Subgroup("/Inventory")
	playfabInventoryGroup.HandleFunc("POST", "/GetInventoryItems", Inventory.GetInventoryItems)

	playfabMultiplayerServerGroup := p.group.Subgroup("/MultiplayerServer")
	playfabMultiplayerServerGroup.HandleFunc("POST", "/GetCognitiveServicesToken", MultiplayerServer.GetCognitiveServicesToken)
	playfabMultiplayerServerGroup.HandleFunc("POST", "/ListPartyQosServers", MultiplayerServer.ListPartyQosServers)

	playfabPartyGroup := p.group.Subgroup("/Party")
	playfabPartyGroup.HandleFunc("POST", "/RequestParty", Party.RequestParty)

	catalogGroup := p.group.Subgroup("/Catalog")
	catalogGroup.HandleFunc("POST", "/GetItems", Catalog.GetItems)

	cloudScriptGroup := p.group.Subgroup("/CloudScript")
	cloudScriptGroup.HandleFunc("POST", "/ExecuteFunction", CloudScript.ExecuteFunction)

	fs := http.FileServer(http.Dir(playfab.BaseDir))
	playfabStaticGroup := p.group.Subgroup(playfab.StaticSuffix)
	playfabStaticGroup.Handle(
		"GET",
		"/",
		http.StripPrefix(playfab.StaticSuffix, fs),
	)
	return PlayfabMiddleware(p.group.mux)
}
