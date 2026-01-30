package models

import (
	"fmt"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	i "github.com/luskaner/ageLANServer/server/internal"
)

var BattleServersStore = make(map[string][]BattleServer)

func InitializeBattleServers(gameId string, configBattleServers []i.BattleServer) error {
	var battleServers []BattleServer
	for _, bs := range configBattleServers {
		battleServers = append(battleServers, &MainBattleServer{
			BaseConfig: bs.BaseConfig,
		})
	}
	tmpBattleServer, err := battleServerConfig.Configs(gameId, true)
	if err != nil {
		return err
	}
	for _, bs := range tmpBattleServer {
		battleServers = append(battleServers, &MainBattleServer{
			BaseConfig: bs.BaseConfig,
		})
	}
	if (gameId == common.GameAoE4 || gameId == common.GameAoM) && len(battleServers) == 0 {
		return fmt.Errorf("no battle server")
	}
	BattleServersStore[gameId] = battleServers
	return nil
}
