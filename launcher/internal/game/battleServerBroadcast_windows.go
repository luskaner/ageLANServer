package game

import "github.com/luskaner/aoe2DELanServer/battle-server-broadcast"

func RequiresBattleServerBroadcast() bool {
	mostPriority, restInterfaces, err := battle_server_broadcast.RetrieveBsInterfaceAddresses()
	if err == nil && mostPriority != nil && len(restInterfaces) > 0 {
		return true
	}
	return false
}
