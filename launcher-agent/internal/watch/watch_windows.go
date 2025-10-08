package watch

import (
	"github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
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

	var event uint32
	event, err = windows.WaitForSingleObject(handle, windows.INFINITE)
	return err == nil && event == uint32(windows.WAIT_OBJECT_0)
}

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
