package watch

import (
	"github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
)

func rebroadcastBattleServer(exitCode *int, port int) {
	mostPriority, restInterfaces, err := battle_server_broadcast.RetrieveBsInterfaceAddresses()
	if err == nil && mostPriority != nil && len(restInterfaces) > 0 {
		if len(waitUntilAnyProcessExist([]string{"BattleServer.exe"})) > 0 {
			go func() {
				_ = battle_server_broadcast.CloneAnnouncements(mostPriority, restInterfaces, port)
			}()
		} else {
			*exitCode = internal.ErrBattleServerTimeOutStart
		}
	}
}
