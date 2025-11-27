package models

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type MainGame struct {
	battleServers  BattleServers
	resources      Resources
	users          Users
	advertisements Advertisements
	chatChannels   ChatChannels
	title          string
}

func CreateMainGame(gameId string, battleServers BattleServers, resources Resources, users Users,
	advertisements Advertisements, chatChannels ChatChannels, rssKeyedFilenames mapset.Set[string],
	battleServerHaveOobPort bool, battleServerName string) Game {
	if battleServers == nil {
		battleServers = &MainBattleServers{}
	}
	if resources == nil {
		resources = &MainResources{}
	}
	if users == nil {
		users = &MainUsers{}
	}
	if advertisements == nil {
		advertisements = &MainAdvertisements{}
	}
	if chatChannels == nil {
		chatChannels = &MainChatChannels{}
	}
	game := &MainGame{
		battleServers:  battleServers,
		resources:      resources,
		users:          users,
		advertisements: advertisements,
		chatChannels:   chatChannels,
		title:          gameId,
	}
	game.battleServers.Initialize(BattleServersStore[gameId], battleServerHaveOobPort, battleServerName)
	game.resources.Initialize(gameId, rssKeyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users, game.battleServers)
	game.chatChannels.Initialize(game.resources.ChatChannels())
	return game
}

func CreateGame(gameId string, rssKeyedFilenames mapset.Set[string], battleServerHaveOobPort bool, battleServerName string) Game {
	return CreateMainGame(gameId, nil, nil, nil, nil, nil,
		rssKeyedFilenames, battleServerHaveOobPort, battleServerName)
}

func (g *MainGame) Resources() Resources {
	return g.resources
}

func (g *MainGame) Users() Users {
	return g.users
}

func (g *MainGame) Advertisements() Advertisements {
	return g.advertisements
}

func (g *MainGame) ChatChannels() ChatChannels {
	return g.chatChannels
}

func (g *MainGame) Title() string {
	return g.title
}

func (g *MainGame) BattleServers() BattleServers {
	return g.battleServers
}
