package models

import (
	"fmt"

	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/game"
	i "github.com/luskaner/ageLANServer/server/internal"
)

var BattleServersStore = make(map[string][]BattleServer)

func InitializeBattleServers(gameId string, configBattleServers []i.BattleServer) error {
	var battleServers []BattleServer
	for _, bs := range configBattleServers {
		battleServers = append(battleServers, &MainBattleServer{
			Base: bs.Base,
		})
	}
	tmpBattleServer, err := battleServer.Configs(gameId, true)
	if err != nil {
		return err
	}
	for _, bs := range tmpBattleServer {
		battleServers = append(battleServers, &MainBattleServer{
			Base: bs.Base,
		})
	}
	if (gameId == game.AoE4 || gameId == game.AoM) && len(battleServers) == 0 {
		return fmt.Errorf("no battle server")
	}
	BattleServersStore[gameId] = battleServers
	return nil
}
