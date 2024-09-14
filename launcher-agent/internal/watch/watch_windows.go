package watch

import (
	"github.com/luskaner/aoe2DELanServer/battle-server-broadcast"
	"github.com/luskaner/aoe2DELanServer/launcher-agent/internal"
	"golang.org/x/sys/windows"
)

func waitForProcess(PID uint32) bool {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, true, PID)

	if err != nil {
		return false
	}

	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(handle)

	_, err = windows.WaitForSingleObject(handle, windows.INFINITE)

	if err != nil {
		return false
	}

	return true
}

func rebroadcastBattleServer(exitCode *int) {
	mostPriority, restInterfaces, err := battle_server_broadcast.RetrieveBsInterfaceAddresses()
	if err == nil && mostPriority != nil && len(restInterfaces) > 0 {
		if len(waitUntilAnyProcessExist([]string{"BattleServer.exe"})) > 0 {
			go func() {
				_ = battle_server_broadcast.CloneAnnouncements(mostPriority, restInterfaces)
			}()
		} else {
			*exitCode = internal.ErrBattleServerTimeOutStart
		}
	}
}
