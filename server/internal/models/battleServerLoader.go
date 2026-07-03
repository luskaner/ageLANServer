package models

import (
	"fmt"

	"github.com/luskaner/ageLANServer/common/battleServer"
	"github.com/luskaner/ageLANServer/common/game"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/logger"
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
	if len(battleServers) == 0 {
		switch gameId {
		case game.AoE2:
			logger.Println("No Battle Server for AoE 2: DE. macOS native clients will not be able to play/observe games.")
		case game.AoE4, game.AoM:
			return fmt.Errorf("no battle server")
		}
	}
	BattleServersStore[gameId] = battleServers
	return nil
}
