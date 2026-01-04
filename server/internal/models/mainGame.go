package models

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type MainGame struct {
	battleServers          BattleServers
	resources              Resources
	users                  Users
	advertisements         Advertisements
	chatChannels           ChatChannels
	sessions               Sessions
	leaderboardDefinitions LeaderboardDefinitions
	title                  string
}

func CreateMainGame(gameId string, battleServers BattleServers, resources Resources, leaderboardDefinitions LeaderboardDefinitions, users Users,
	advertisements Advertisements, chatChannels ChatChannels, sessions Sessions, rssKeyedFilenames mapset.Set[string],
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
	if sessions == nil {
		sessions = &MainSessions{}
	}
	game := &MainGame{
		battleServers:  battleServers,
		resources:      resources,
		users:          users,
		advertisements: advertisements,
		chatChannels:   chatChannels,
		sessions:       sessions,
		title:          gameId,
	}
	game.battleServers.Initialize(BattleServersStore[gameId], battleServerHaveOobPort, battleServerName)
	game.resources.Initialize(gameId, rssKeyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users, game.battleServers)
	game.chatChannels.Initialize(game.resources.ChatChannels())
	game.sessions.Initialize()
	if leaderboards, ok := game.resources.ArrayFiles()["leaderboards.json"]; ok {
		if leaderboardDefinitions == nil {
			leaderboardDefinitions = &MainLeaderboardDefinitions{}
		}
		game.leaderboardDefinitions = leaderboardDefinitions
		game.leaderboardDefinitions.Initialize(leaderboards)
	}
	return game
}

func CreateGame(gameId string, rssKeyedFilenames mapset.Set[string], battleServerHaveOobPort bool, battleServerName string) Game {
	return CreateMainGame(gameId, nil, nil, nil, nil, nil, nil,
		nil, rssKeyedFilenames, battleServerHaveOobPort, battleServerName)
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

func (g *MainGame) Sessions() Sessions {
	return g.sessions
}

func (g *MainGame) LeaderboardDefinitions() LeaderboardDefinitions {
	return g.leaderboardDefinitions
}
