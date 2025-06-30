package watch

import (
	"github.com/luskaner/ageLANServer/battle-server-broadcast"
	"github.com/luskaner/ageLANServer/launcher-agent/internal"
	"golang.org/x/sys/windows"
	"net"
	"slices"
	"time"
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
	event, err = windows.WaitForSingleObject(handle, uint32((5 * time.Minute).Milliseconds()))

	if err != nil || event == uint32(windows.WAIT_TIMEOUT) {
		return false
	}

	return true
}

func rebroadcastBattleServer(exitCode *int, rebroadcastIPs []net.IP, port int) {
	mostPriority, restInterfaces, err := battleServerBroadcast.RetrieveBsInterfaceAddresses()
	if err == nil && mostPriority != nil && len(restInterfaces) > 0 {
		var selectedInterfaces []*net.IPNet
		for _, iff := range restInterfaces {
			if slices.ContainsFunc(rebroadcastIPs, func(ip net.IP) bool {
				return ip.Equal(iff.IP)
			}) {
				selectedInterfaces = append(selectedInterfaces, iff)
			}
		}
		if len(selectedInterfaces) > 0 {
			if len(waitUntilAnyProcessExist([]string{"BattleServer.exe"})) > 0 {
				go func() {
					_ = battleServerBroadcast.CloneAnnouncements(mostPriority, selectedInterfaces, port)
				}()
			} else {
				*exitCode = internal.ErrBattleServerTimeOutStart
			}
		}
	}
}
