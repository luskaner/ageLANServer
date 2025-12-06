package athens

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/game/communityEvent"
	"github.com/luskaner/ageLANServer/server/internal/models/athens/playfab"
	commonPlayfab "github.com/luskaner/ageLANServer/server/internal/models/playfab"
)

type Game struct {
	models.Game
	Blessings    []playfab.Blessings
	CatalogItems map[string]commonPlayfab.CatalogItem
	// All users have the same fixed items
	InventoryItems []commonPlayfab.InventoryItem
}

func CreateGame() models.Game {
	mainGame := models.CreateMainGame(
		common.GameAoM,
		nil,
		nil,
		nil,
		nil,
		nil,
		mapset.NewThreadUnsafeSet[string]("itemBundleItems.json", "itemDefinitions.json"),
		true,
		"true",
	)
	g := &Game{Game: mainGame, Blessings: playfab.ReadBlessings()}
	g.CatalogItems, g.InventoryItems = playfab.Items(g.Blessings)
	communityEvent.Initialize()
	return g
}

func (g *Game) CommunityEventsEncoded() internal.A {
	return communityEvent.CommunityEventsEncoded()
}
