package models

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type MainGame struct {
	battleServers  *MainBattleServers
	resources      *MainResources
	users          *MainUsers
	advertisements *MainAdvertisements
	chatChannels   *MainChatChannels
	title          string
}

func CreateMainGame(gameId string, rssKeyedFilenames mapset.Set[string], battleServerHaveOobPort bool, battleServerName string) *MainGame {
	game := &MainGame{
		battleServers:  &MainBattleServers{},
		resources:      &MainResources{},
		users:          &MainUsers{},
		advertisements: &MainAdvertisements{},
		chatChannels:   &MainChatChannels{},
		title:          gameId,
	}
	game.battleServers.Initialize(BattleServers[gameId], battleServerHaveOobPort, battleServerName)
	game.resources.Initialize(gameId, rssKeyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users, game.battleServers)
	game.chatChannels.Initialize(game.resources.ChatChannels)
	return game
}

func CreateGame(gameId string, rssKeyedFilenames mapset.Set[string], battleServerHaveOobPort bool, battleServerName string) Game {
	return CreateMainGame(gameId, rssKeyedFilenames, battleServerHaveOobPort, battleServerName)
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
