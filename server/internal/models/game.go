package models

import "net/http"

type Game interface {
	Title() string
	Resources() Resources
	Items() Items
	LeaderboardDefinitions() LeaderboardDefinitions
	PresenceDefinitions() PresenceDefinitions
	BattleServers() BattleServers
	Users() Users
	Advertisements() Advertisements
	ChatChannels() ChatChannels
	Sessions() Sessions
}

func G(r *http.Request) Game {
	return Gg[Game](r)
}

func Gg[T Game](r *http.Request) T {
	return r.Context().Value("game").(T)
}

type MainGame struct {
	battleServers          BattleServers
	resources              Resources
	users                  Users
	advertisements         Advertisements
	chatChannels           ChatChannels
	sessions               Sessions
	leaderboardDefinitions LeaderboardDefinitions
	items                  Items
	presenceDefinitions    PresenceDefinitions
	title                  string
}

type CreateMainGameOpts struct {
	Resources    *ResourcesOpts
	BattleServer *BattleServerOpts
	Instances    *InstanceOpts
}

type InstanceOpts struct {
	Resources              Resources
	BattleServers          BattleServers
	Users                  Users
	Advertisements         Advertisements
	ChatChannels           ChatChannels
	Sessions               Sessions
	LeaderboardDefinitions LeaderboardDefinitions
	PresenceDefinitions    PresenceDefinitions
	Items                  Items
}

func CreateMainGame(gameId string, opts *CreateMainGameOpts) Game {
	if opts == nil {
		opts = &CreateMainGameOpts{}
	}
	if opts.Resources == nil {
		opts.Resources = &ResourcesOpts{}
	}
	if opts.Instances == nil {
		opts.Instances = &InstanceOpts{}
	}
	if opts.Instances.BattleServers == nil {
		opts.Instances.BattleServers = &MainBattleServers{}
	}
	if opts.Instances.Resources == nil {
		opts.Instances.Resources = &MainResources{}
	}
	if opts.Instances.Users == nil {
		opts.Instances.Users = &MainUsers{}
	}
	if opts.Instances.Advertisements == nil {
		opts.Instances.Advertisements = &MainAdvertisements{}
	}
	if opts.Instances.ChatChannels == nil {
		opts.Instances.ChatChannels = &MainChatChannels{}
	}
	if opts.Instances.Sessions == nil {
		opts.Instances.Sessions = &MainSessions{}
	}
	game := &MainGame{
		battleServers:  opts.Instances.BattleServers,
		resources:      opts.Instances.Resources,
		users:          opts.Instances.Users,
		advertisements: opts.Instances.Advertisements,
		chatChannels:   opts.Instances.ChatChannels,
		sessions:       opts.Instances.Sessions,
		title:          gameId,
	}
	game.battleServers.Initialize(BattleServersStore[gameId], opts.BattleServer)
	game.resources.Initialize(gameId, opts.Resources)
	game.users.Initialize()
	game.advertisements.Initialize(game.users, game.battleServers)
	game.chatChannels.Initialize(game.resources.ChatChannels())
	game.sessions.Initialize()
	if itemLocations, ok := game.resources.ArrayFiles()["itemLocations.json"]; ok {
		if opts.Instances.Items == nil {
			opts.Instances.Items = &MainItems{}
		}
		game.items = opts.Instances.Items
		itemDefinitions, _ := game.resources.SignedAssets()["itemDefinitions.json"]
		game.items.Initialize(itemDefinitions, itemLocations)
	}
	if leaderboards, ok := game.resources.ArrayFiles()["leaderboards.json"]; ok {
		if opts.Instances.LeaderboardDefinitions == nil {
			opts.Instances.LeaderboardDefinitions = &MainLeaderboardDefinitions{}
		}
		game.leaderboardDefinitions = opts.Instances.LeaderboardDefinitions
		game.leaderboardDefinitions.Initialize(leaderboards)
	}
	if presenceDefinitions, ok := game.resources.ArrayFiles()["presenceData.json"]; ok {
		if opts.Instances.PresenceDefinitions == nil {
			opts.Instances.PresenceDefinitions = &MainPresenceDefinitions{}
		}
		game.presenceDefinitions = opts.Instances.PresenceDefinitions
		game.presenceDefinitions.Initialize(presenceDefinitions)
	}
	return game
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

func (g *MainGame) Items() Items {
	return g.items
}

func (g *MainGame) PresenceDefinitions() PresenceDefinitions {
	return g.presenceDefinitions
}
