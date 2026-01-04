package internal

import "github.com/luskaner/ageLANServer/common/battleServerConfig"

type Announcement struct {
	Enabled        bool
	Multicast      bool
	MulticastGroup string
	Port           int
}

type BattleServer struct {
	battleServerConfig.BaseConfig `mapstructure:",squash"`
}

type Game struct {
	Hosts         []string
	BattleServers []BattleServer
}

type Games struct {
	Enabled []string
	Age1    Game
	Age2    Game
	Age3    Game
	Athens  Game
}

type Configuration struct {
	Log                    bool
	GeneratePlatformUserId bool
	Announcement           Announcement
	Games                  Games
}

func (cfg *Configuration) GetGameHosts(gameId string) []string {
	switch gameId {
	case "age1":
		return cfg.Games.Age1.Hosts
	case "age2":
		return cfg.Games.Age2.Hosts
	case "age3":
		return cfg.Games.Age3.Hosts
	case "athens":
		return cfg.Games.Athens.Hosts
	default:
		return nil
	}
}

func (cfg *Configuration) GetGameBattleServers(gameId string) []BattleServer {
	switch gameId {
	case "age1":
		return cfg.Games.Age1.BattleServers
	case "age2":
		return cfg.Games.Age2.BattleServers
	case "age3":
		return cfg.Games.Age3.BattleServers
	case "athens":
		return cfg.Games.Athens.BattleServers
	default:
		return nil
	}
}
