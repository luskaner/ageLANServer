package models

import (
	"fmt"
	"os"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/luskaner/ageLANServer/server/internal"
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
	tmpBattleServer, err := battleServerConfig.Configs(gameId, true)
	if err != nil {
		panic(err)
	}
	for _, bs := range tmpBattleServer {
		battleServers = append(battleServers, MainBattleServer{
			BaseConfig: bs.BaseConfig,
		})
	}
	if gameId == common.GameAoM && len(battleServers) == 0 {
		fmt.Println("AoM: RT requires a Battle Server. You can start one with 'battle-server-manager'.")
		os.Exit(internal.MissingBattleServer)
	}
	if len(battleServers) > 0 {
		fmt.Printf("Battle Servers for %s:\n", gameId)
	}
	game.battleServers.Initialize(battleServers, battleServerHaveOobPort, battleServerName)
	for _, bs := range game.battleServers.Iter() {
		fmt.Println(bs.String())
	}
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
