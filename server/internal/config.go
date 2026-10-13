package internal

import (
	"github.com/luskaner/ageLANServer/common/battleServer"
)

type Announcement struct {
	Enabled        bool
	Multicast      bool
	MulticastGroup string
	Port           int
}

type BattleServer struct {
	battleServer.Base `koanf:",squash"`
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
	Age4    Game
	Athens  Game
}

type Internet struct {
	Enabled bool
	IP      string
}

// CustomFiles configures the /api/MatchContent/ endpoints (multiplayer
// savegame and replay upload/download). Sizes and the TTL are kept as strings
// and validated once at startup, so a typo fails the boot instead of the first
// upload.
type CustomFiles struct {
	Enabled      bool
	Folder       string
	MaxFileSize  string
	MaxTotalSize string
	Expiry       string
}

type Configuration struct {
	Log                    bool
	GeneratePlatformUserId bool
	Authentication         string
	Internet               Internet
	Announcement           Announcement
	Games                  Games
	CustomFiles            CustomFiles
	// Sign makes the server re-sign the catalogues it serves instead of
	// replaying the signatures captured alongside them in
	// resources/responses. Off by default: the captured ones are what an
	// unmodified client accepts.
	Sign bool
}

func (cfg *Configuration) GetGameHosts(gameId string) []string {
	switch gameId {
	case "age1":
		return cfg.Games.Age1.Hosts
	case "age2":
		return cfg.Games.Age2.Hosts
	case "age3":
		return cfg.Games.Age3.Hosts
	case "age4":
		return cfg.Games.Age4.Hosts
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
	case "age4":
		return cfg.Games.Age4.BattleServers
	case "athens":
		return cfg.Games.Athens.BattleServers
	default:
		return nil
	}
}
