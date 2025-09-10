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

func CreateGame(gameId string, rssKeyedFilenames mapset.Set[string], battleServerHaveOobPort bool, battleServerName string) Game {
	game := &MainGame{
		battleServers:  &MainBattleServers{},
		resources:      &MainResources{},
		users:          &MainUsers{},
		advertisements: &MainAdvertisements{},
		chatChannels:   &MainChatChannels{},
		title:          gameId,
	}
	var battleServers []MainBattleServer
	key := fmt.Sprintf("Games.%s.BattleServers", gameId)
	if viper.IsSet(key) {
		err := viper.UnmarshalKey(key, &battleServers)
		if err != nil {
			panic(err)
		}
	}
	game.battleServers.Initialize(battleServers, battleServerHaveOobPort, battleServerName)
	game.resources.Initialize(gameId, rssKeyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users, game.battleServers)
	game.chatChannels.Initialize(game.resources.ChatChannels)
	return game
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

func (g *MainGame) BattleServers() *MainBattleServers {
	return g.battleServers
}
