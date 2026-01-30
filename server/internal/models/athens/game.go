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
	commonPlayfab.BaseGame
	AllowedBlessings              map[int][]string
	GauntletPoolIndexByDifficulty map[string][]int
	Gauntlet                      playfab.Gauntlet
	GauntletMissionPools          playfab.GauntletMissionPools
	CatalogItems                  map[string]commonPlayfab.CatalogItem
	// All users have the same fixed items
	InventoryItems []commonPlayfab.InventoryItem
}

func CreateGame() models.Game {
	mainGame := models.CreateMainGame(
		common.GameAoM,
		&models.CreateMainGameOpts{
			Instances: &models.InstanceOpts{
				Users: &user.Users{},
			},
			Resources: &models.ResourcesOpts{
				KeyedFilenames: mapset.NewThreadUnsafeSet[string]("itemBundleItems.json", "itemDefinitions.json"),
			},
		},
	)
	g := &Game{
		BaseGame: commonPlayfab.BaseGame{
			Game: mainGame,
		},
	}
	blessings := playfab.ReadBlessings()
	g.CatalogItems, g.InventoryItems = playfab.Items(blessings)
	g.Gauntlet = playfab.ReadGauntlet()
	g.GauntletMissionPools = playfab.ReadGauntletMissionPools()
	g.AllowedBlessings = precomputed.AllowedGauntletBlessings(g.Gauntlet, blessings)
	gauntletPoolNamesToIndex := precomputed.PoolNamesToIndex(g.GauntletMissionPools)
	g.GauntletPoolIndexByDifficulty = precomputed.PoolsIndexByDifficulty(g.Gauntlet, gauntletPoolNamesToIndex)
	g.PlayfabSessions().Initialize()
	communityEvent.Initialize()
	return g
}

func (g *Game) CommunityEventsEncoded() internal.A {
	return communityEvent.CommunityEventsEncoded()
}
