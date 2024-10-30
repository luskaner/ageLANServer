package models

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type MainGame struct {
	resources      *MainResources
	users          *MainUsers
	advertisements *MainAdvertisements
	chatChannels   *MainChatChannels
	title          string
}

func CreateGame(gameId string, rssKeyedFilenames mapset.Set[string]) Game {
	game := &MainGame{
		resources:      &MainResources{},
		users:          &MainUsers{},
		advertisements: &MainAdvertisements{},
		chatChannels:   &MainChatChannels{},
		title:          gameId,
	}
	game.resources.Initialize(gameId, rssKeyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users)
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
