package models

import (
	"fmt"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/battleServerConfig"
	"github.com/spf13/viper"
)

var BattleServers = make(map[string][]*MainBattleServer)

func InitializeBattleServers(gameId string) error {
	var battleServers []*MainBattleServer
	key := fmt.Sprintf("Games.%s.BattleServers", gameId)
	if viper.IsSet(key) {
		err := viper.UnmarshalKey(key, &battleServers)
		if err != nil {
			return err
		}
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
	if gameId == common.GameAoM && len(battleServers) == 0 {
		return fmt.Errorf("no battle server for AoM")
	}
	BattleServers[gameId] = battleServers
	return nil
}
