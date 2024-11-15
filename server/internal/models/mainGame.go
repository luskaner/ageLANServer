package models

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/viper"
)

type MainGame struct {
	battleServers  *MainBattleServers
	resources      *MainResources
	users          *MainUsers
	advertisements *MainAdvertisements
	chatChannels   *MainChatChannels
	title          string
}

func CreateGame(gameId string, rssKeyedFilenames mapset.Set[string], battleServersEncodeOobPort bool) Game {
	game := &MainGame{
		battleServers:  &MainBattleServers{},
		resources:      &MainResources{},
		users:          &MainUsers{},
		advertisements: &MainAdvertisements{},
		chatChannels:   &MainChatChannels{},
		title:          gameId,
	}
	game.battleServers.Initialize(
		viper.GetStringMap(fmt.Sprintf("BattleServers.%s", gameId)), battleServersEncodeOobPort,
	)
	game.resources.Initialize(gameId, rssKeyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users)
	game.chatChannels.Initialize(game.resources.ChatChannels)
	return game
}

func (g *MainGame) BattleServers() *MainBattleServers {
	return g.battleServers
}

func (g *MainGame) Resources() *MainResources {
	return g.resources
}

func (g *MainGame) Users() *MainUsers {
	return g.users
}

func (g *MainGame) Advertisements() *MainAdvertisements {
	return g.advertisements
}

func (g *MainGame) ChatChannels() *MainChatChannels {
	return g.chatChannels
}

func (g *MainGame) Title() string {
	return g.title
}
