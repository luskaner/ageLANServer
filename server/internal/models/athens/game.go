package athens

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/routes/game/communityEvent"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/routes/playfab"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/routes/playfab/cloudScriptFunction/buildGauntletLabyrinth/precomputed"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/user"
	commonPlayfab "github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

type Game struct {
	models.Game
	AllowedBlessings              map[int][]string
	GauntletPoolIndexByDifficulty map[string][]int
	Gauntlet                      playfab.Gauntlet
	GauntletMissionPools          playfab.GauntletMissionPools
	CatalogItems                  map[string]commonPlayfab.CatalogItem
	// All users have the same fixed items
	InventoryItems  []commonPlayfab.InventoryItem
	PlayfabSessions commonPlayfab.MainSessions
}

func CreateGame() models.Game {
	mainGame := models.CreateMainGame(
		common.GameAoM,
		nil,
		nil,
		&user.Users{},
		nil,
		nil,
		nil,
		mapset.NewThreadUnsafeSet[string]("itemBundleItems.json", "itemDefinitions.json"),
		true,
		"true",
	)
	g := &Game{
		Game: mainGame,
	}
	blessings := playfab.ReadBlessings()
	g.CatalogItems, g.InventoryItems = playfab.Items(blessings)
	g.Gauntlet = playfab.ReadGauntlet()
	g.GauntletMissionPools = playfab.ReadGauntletMissionPools()
	g.AllowedBlessings = precomputed.AllowedGauntletBlessings(g.Gauntlet, blessings)
	gauntletPoolNamesToIndex := precomputed.PoolNamesToIndex(g.GauntletMissionPools)
	g.GauntletPoolIndexByDifficulty = precomputed.PoolsIndexByDifficulty(g.Gauntlet, gauntletPoolNamesToIndex)
	g.PlayfabSessions.Initialize()
	communityEvent.Initialize()
	return g
}

func (g *Game) CommunityEventsEncoded() internal.A {
	return communityEvent.CommunityEventsEncoded()
}
